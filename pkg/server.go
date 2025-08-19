package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

const (
	pluginApi              = "https://registry.opentofu.org/v1/providers"
	providerFileNamePrefix = "terraform-provider-"
	urlPathSeparator       = '/'
)

var (
	ErrPluginNotFound = fmt.Errorf("plugin not found")
	ErrPluginApi      = fmt.Errorf("plugin API error")
)

// ContextKey is a type used to store the server instance in the context.
type ContextKey struct{}

// Request is a request structure used to specify the details of a plugin
// so that it can be downloaded.
// Note that the request fields are case-sensitive.
type Request struct {
	Namespace string // Namespace of the provider (e.g., "Azure")
	Name      string // Name of the provider (e.g., "azapi")
	Version   string // Version of the provider (e.g., "2.5.0")
}

// String returns a string representation of the Request in the format:
// "https://registry.opentofu.org/v1/providers/{namespace}/{name}/{version}/download/{os}/{arch}"
// This format is used to construct the URL for downloading the plugin.
func (r Request) String() string {
	sb := strings.Builder{}
	sb.WriteString(pluginApi)
	sb.WriteRune(urlPathSeparator)
	sb.WriteString(r.Namespace)
	sb.WriteRune(urlPathSeparator)
	sb.WriteString(r.Name)
	sb.WriteRune(urlPathSeparator)
	sb.WriteString(r.Version)
	sb.WriteString("/download/")
	sb.WriteString(runtime.GOOS)
	sb.WriteRune(urlPathSeparator)
	sb.WriteString(runtime.GOARCH)
	result := sb.String()
	if _, err := url.Parse(result); err != nil {
		panic(fmt.Sprintf("failed to parse URL: %s, error: %v", result, err))
	}
	return result
}

type pluginApiResponse struct {
	Protocols   []string `json:"protocols"`
	OS          string   `json:"os"`
	Arch        string   `json:"arch"`
	FileName    string   `json:"filename"`
	DownloadURL string   `json:"download_url"`
}

type downloadCache map[Request]string
type schemaCache map[Request]*tfjson.ProviderSchema

// Server is a struct that manages the plugin download and caching process.
type Server struct {
	tmpDir string
	dlc    downloadCache
	sc     schemaCache
	l      *slog.Logger
}

// NewServer creates a new Server instance with an optional logger.
// If no logger is provided, it defaults to a logger that discards all logs.
func NewServer(l *slog.Logger) *Server {
	if l == nil {
		l = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level:     slog.LevelError,
			AddSource: false,
		}))
	}
	l.Info("Creating new server instance")
	return &Server{
		dlc: make(downloadCache),
		sc:  make(schemaCache),
		l:   l,
	}
}

// Cleanup removes the temporary directory used for plugin downloads.
func (s *Server) Cleanup() {
	s.l.Info("Cleaning up temporary directory", "dir", s.tmpDir)
	os.RemoveAll(s.tmpDir)
}

// Get retrieves the plugin for the specified request, downloading it if necessary.
// The GetXxx methods (GetResourceSchema, GetDataSourceSchema, etc.) will call this method anyway,
// so it is not necessary to call Get directly unless you want to ensure the plugin is downloaded first.
// It is stored in a temporary directory and cached for future use.
// Make sure to call Cleanup() to remove the temporary files.
func (s *Server) Get(request Request) error {
	l := s.l.With("request_namespace", request.Namespace, "request_name", request.Name, "request_version", request.Version)
	if _, exists := s.dlc[request]; exists {
		l.Info("Request already exists in download cache")
		return nil // Request already exists, no need to add again
	}

	registryApiRequest, err := http.NewRequest(http.MethodGet, request.String(), nil)
	l.Debug("Sending request to registry API", "url", registryApiRequest.URL.String())
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for registry API: %w", err)
	}

	resp, err := http.DefaultClient.Do(registryApiRequest)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request to registry API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s", ErrPluginNotFound, request.String())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s => %d", ErrPluginApi, request.String(), resp.StatusCode)
	}

	var pluginResponse pluginApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&pluginResponse); err != nil {
		return fmt.Errorf("failed to decode plugin API response: %w", err)
	}

	l.Info("Plugin API response received", "arch", pluginResponse.Arch, "os", pluginResponse.OS, "filename", pluginResponse.FileName, "download_url", pluginResponse.DownloadURL)

	if s.tmpDir == "" {
		tmpFile, err := os.MkdirTemp("", "tfpluginschema-")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		s.tmpDir = tmpFile
	}

	downloadURL := pluginResponse.DownloadURL
	if downloadURL == "" {
		return fmt.Errorf("download URL is empty for request: %s", request.String())
	}

	downloadRequest, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for plugin download: %w", err)
	}

	resp, err = http.DefaultClient.Do(downloadRequest)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download plugin: %s => %d", downloadURL, resp.StatusCode)
	}

	pluginFilePath := filepath.Join(s.tmpDir, pluginResponse.FileName)

	file, err := os.Create(pluginFilePath)
	if err != nil {
		return fmt.Errorf("failed to create plugin file: %w", err)
	}

	defer file.Close()

	if _, err := file.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("failed to read plugin data into file: %w", err)
	}

	// unzip the file
	extractDir := strings.TrimSuffix(pluginResponse.FileName, filepath.Ext(pluginResponse.FileName)) // Remove extension for directory name
	extractDir = filepath.Join(s.tmpDir, extractDir)

	if err := os.Mkdir(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := unzip(pluginFilePath, extractDir); err != nil {
		return fmt.Errorf("failed to unzip plugin file: %w", err)
	}

	// check the extracted directory
	wantProviderFileName := fmt.Sprintf("%s%s", providerFileNamePrefix, request.Name)
	if err = fs.WalkDir(os.DirFS(extractDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking extracted directory (%s): %w", extractDir, err)
		}

		l.Debug("Checking extracted file", "path", path, "is_dir", d.IsDir(), "name", d.Name())

		if d.IsDir() && d.Name() != "." {
			return fs.SkipDir
		}

		if !strings.HasPrefix(d.Name(), wantProviderFileName) {
			return nil
		}

		l.Info("Found provider file", "provider_file_name", d.Name())

		s.dlc[request] = filepath.Join(extractDir, path)
		return fs.SkipAll
	}); err != nil {
		return fmt.Errorf("error checking extracted files: %w", err)
	}

	if _, exists := s.dlc[request]; !exists {
		return fmt.Errorf("provider file not found in extracted directory (%s) for request: %s", extractDir, request.String())
	}

	return nil
}

// getSchema creates a universal provider client for the given request
func (s *Server) getSchema(request Request) ([]byte, error) {
	if resp, exists := s.sc[request]; exists {
		return json.Marshal(resp)
	}

	// Ensure the provider is downloaded
	if err := s.Get(request); err != nil {
		return nil, fmt.Errorf("failed to download provider: %w", err)
	}

	// Get the provider path
	providerPath, exists := s.dlc[request]
	if !exists {
		return nil, fmt.Errorf("provider not found in cache: %s", request.String())
	}

	client, err := newGrpcClient(providerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}
	defer client.close()

	var schemaResp *tfjson.ProviderSchema

	// Try V6 protocol first
	if resp, v6Err := client.v6Schema(); v6Err == nil {
		schemaResp, err = convertV6ResponseToProviderSchema(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert V6 provider schema: %w", err)
		}
	} else if resp, v5Err := client.v5Schema(); v5Err == nil {
		// Fall back to V5 protocol
		schemaResp, err = convertV5ResponseToProviderSchema(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert V5 provider schema: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to get provider schema for either V5 or V6 protocols: v6 error: %v, v5 error: %v", v6Err, v5Err)
	}

	// Cache the schema response
	s.sc[request] = schemaResp

	return json.MarshalIndent(schemaResp, "", "  ")
}

// GetResourceSchema retrieves the schema for a specific resource from the provider.
func (s *Server) GetResourceSchema(request Request, resource string) (*tfjson.Schema, error) {
	s.l.Info("Getting resource schema", "request", request, "resource", resource)
	schemaResp, ok := s.sc[request]
	if !ok {
		if _, err := s.getSchema(request); err != nil {
			return nil, fmt.Errorf("failed to read provider schema: %w", err)
		}
		schemaResp = s.sc[request]
	}

	schemaResource, ok := schemaResp.ResourceSchemas[resource]
	if !ok {
		return nil, fmt.Errorf("resource schema not found: %s", resource)
	}
	return schemaResource, nil
}

// GetDataSourceSchema retrieves the schema for a specific data source from the provider.
func (s *Server) GetDataSourceSchema(request Request, dataSource string) (*tfjson.Schema, error) {
	s.l.Info("Getting data source schema", "request", request, "data_source", dataSource)
	schemaResp, ok := s.sc[request]
	if !ok {
		if _, err := s.getSchema(request); err != nil {
			return nil, fmt.Errorf("failed to read provider schema: %w", err)
		}
		schemaResp = s.sc[request]
	}

	schemaResource, ok := schemaResp.DataSourceSchemas[dataSource]
	if !ok {
		return nil, fmt.Errorf("data source schema not found: %s", dataSource)
	}
	return schemaResource, nil
}

// GetFunctionSchema retrieves the schema for a specific function from the provider.
func (s *Server) GetFunctionSchema(request Request, function string) (*tfjson.FunctionSignature, error) {
	s.l.Info("Getting function schema", "request", request, "function", function)
	schemaResp, ok := s.sc[request]
	if !ok {
		if _, err := s.getSchema(request); err != nil {
			return nil, fmt.Errorf("failed to read provider schema: %w", err)
		}
		schemaResp = s.sc[request]
	}

	schemaFunction, ok := schemaResp.Functions[function]
	if !ok {
		return nil, fmt.Errorf("function schema not found: %s", function)
	}

	return schemaFunction, nil
}

// GetEphemeralResourceSchema retrieves the schema for a specific ephemeral resource from the provider.
func (s *Server) GetEphemeralResourceSchema(request Request, ephemeralResource string) (*tfjson.Schema, error) {
	s.l.Info("Getting ephemeral resource schema", "request", request, "ephemeral_resource", ephemeralResource)
	schemaResp, ok := s.sc[request]
	if !ok {
		if _, err := s.getSchema(request); err != nil {
			return nil, fmt.Errorf("failed to read provider schema: %w", err)
		}
		schemaResp = s.sc[request]
	}

	schemaResource, ok := schemaResp.EphemeralResourceSchemas[ephemeralResource]
	if !ok {
		return nil, fmt.Errorf("ephemeral resource schema not found: %s", ephemeralResource)
	}

	return schemaResource, nil
}

// GetProviderSchema retrieves the schema for the provider configuration.
func (s *Server) GetProviderSchema(request Request) (*tfjson.ProviderSchema, error) {
	s.l.Info("Getting provider schema", "request", request)
	schemaResp, ok := s.sc[request]
	if !ok {
		if _, err := s.getSchema(request); err != nil {
			return nil, fmt.Errorf("failed to read provider schema: %w", err)
		}
		schemaResp = s.sc[request]
	}

	return schemaResp, nil
}

// marshalResponse marshals the response into JSON and compacts it for better suitability with LLMs.
func marshalResponse(resp any) ([]byte, error) {
	marshalled, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	var compacted bytes.Buffer
	json.Compact(&compacted, marshalled)
	return compacted.Bytes(), nil
}
