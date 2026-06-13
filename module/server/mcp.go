package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/gwaysys/bcstorage/module/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Input / Output Types ----

type ReadFileInput struct {
	Space  string `json:"space" jsonschema:"the user space/namespace"`
	Path   string `json:"path" jsonschema:"file path relative to space"`
	Offset int64  `json:"offset,omitempty" jsonschema:"read offset in bytes (default 0)"`
	Size   int64  `json:"size,omitempty" jsonschema:"max bytes to read (default entire file)"`
}

type ReadFileOutput struct {
	Content string `json:"content" jsonschema:"file content"`
	Size    int64  `json:"size" jsonschema:"total file size in bytes"`
}

type WriteFileInput struct {
	Space   string `json:"space" jsonschema:"the user space/namespace"`
	Path    string `json:"path" jsonschema:"file path relative to space"`
	Content string `json:"content" jsonschema:"file content to write"`
	Offset  int64  `json:"offset,omitempty" jsonschema:"write offset in bytes (default 0)"`
}

type WriteFileOutput struct {
	Written int64 `json:"written" jsonschema:"bytes written"`
}

type DeleteFileInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
	Path  string `json:"path" jsonschema:"file path relative to space"`
}

type DeleteFileOutput struct {
	Deleted bool `json:"deleted" jsonschema:"whether the file was deleted"`
}

type MoveFileInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
	From  string `json:"from" jsonschema:"source path relative to space"`
	To    string `json:"to" jsonschema:"destination path relative to space"`
}

type MoveFileOutput struct {
	Moved bool `json:"moved" jsonschema:"whether the file was moved"`
}

type ListDirInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
	Path  string `json:"path" jsonschema:"directory path relative to space"`
}

type FileEntry struct {
	Name    string `json:"name" jsonschema:"file or directory name"`
	IsDir   bool   `json:"is_dir" jsonschema:"whether this entry is a directory"`
	Size    int64  `json:"size" jsonschema:"file size in bytes"`
	ModTime string `json:"mod_time" jsonschema:"last modification time (RFC3339)"`
}

type ListDirOutput struct {
	Entries []FileEntry `json:"entries" jsonschema:"directory entries"`
}

type StatFileInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
	Path  string `json:"path" jsonschema:"file path relative to space"`
}

type StatFileOutput struct {
	Name    string `json:"name" jsonschema:"file or directory name"`
	IsDir   bool   `json:"is_dir" jsonschema:"whether this is a directory"`
	Size    int64  `json:"size" jsonschema:"file size in bytes"`
	ModTime string `json:"mod_time" jsonschema:"last modification time (RFC3339)"`
}

type TruncateFileInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
	Path  string `json:"path" jsonschema:"file path relative to space"`
	Size  int64  `json:"size" jsonschema:"target file size in bytes"`
}

type TruncateFileOutput struct {
	Truncated bool `json:"truncated" jsonschema:"whether the file was truncated"`
}

type GetCapacityInput struct {
	Space string `json:"space" jsonschema:"the user space/namespace"`
}

type CapacityInfo struct {
	Total     int64 `json:"total" jsonschema:"total space in bytes"`
	Available int64 `json:"available" jsonschema:"available space in bytes"`
	Used      int64 `json:"used" jsonschema:"used space in bytes"`
	Free      int64 `json:"free" jsonschema:"free space in bytes"`
}

// ---- Path Helpers ----

func resolveSpacePath(space string) (string, error) {
	sp, ok := _userMap.GetSpace(space)
	if !ok {
		return "", errors.ErrNoData.As("space not found: " + space)
	}
	return filepath.Join(_dataPath, sp.Name), nil
}

func resolveFilePath(space, filePath string) (string, error) {
	spacePath, err := resolveSpacePath(space)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filepath.Join(spacePath, filePath))
	if err != nil {
		return "", errors.As(err)
	}
	if !strings.HasPrefix(absPath, spacePath) {
		return "", errors.New("path is outside space")
	}
	return absPath, nil
}

// ---- Tool Handlers ----

func readFileHandler(ctx context.Context, req *mcp.CallToolRequest, input ReadFileInput) (*mcp.CallToolResult, ReadFileOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, ReadFileOutput{}, errors.As(err)
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, ReadFileOutput{}, errors.As(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, ReadFileOutput{}, errors.As(err)
	}

	if input.Offset > 0 {
		if _, err := f.Seek(input.Offset, 0); err != nil {
			return nil, ReadFileOutput{}, errors.As(err)
		}
	}

	readSize := stat.Size() - input.Offset
	if input.Size > 0 && input.Size < readSize {
		readSize = input.Size
	}

	if readSize <= 0 {
		return nil, ReadFileOutput{Content: "", Size: stat.Size()}, nil
	}

	buf := make([]byte, readSize)
	n, err := f.Read(buf)
	if err != nil {
		return nil, ReadFileOutput{}, errors.As(err)
	}

	return nil, ReadFileOutput{Content: string(buf[:n]), Size: stat.Size()}, nil
}

func writeFileHandler(ctx context.Context, req *mcp.CallToolRequest, input WriteFileInput) (*mcp.CallToolResult, WriteFileOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, WriteFileOutput{}, errors.As(err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, WriteFileOutput{}, errors.As(err)
	}

	toFile, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, WriteFileOutput{}, errors.As(err)
	}
	defer toFile.Close()

	if input.Offset > 0 {
		if _, err := toFile.Seek(input.Offset, 0); err != nil {
			return nil, WriteFileOutput{}, errors.As(err)
		}
	}

	n, err := toFile.WriteString(input.Content)
	if err != nil {
		return nil, WriteFileOutput{}, errors.As(err)
	}

	log.Infof("MCP write file %s, offset:%d, size:%d", absPath, input.Offset, n)
	return nil, WriteFileOutput{Written: int64(n)}, nil
}

func deleteFileHandler(ctx context.Context, req *mcp.CallToolRequest, input DeleteFileInput) (*mcp.CallToolResult, DeleteFileOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, DeleteFileOutput{}, errors.As(err)
	}

	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			return nil, DeleteFileOutput{Deleted: false}, nil
		}
		return nil, DeleteFileOutput{}, errors.As(err)
	}

	log.Infof("MCP delete file %s", absPath)
	return nil, DeleteFileOutput{Deleted: true}, nil
}

func moveFileHandler(ctx context.Context, req *mcp.CallToolRequest, input MoveFileInput) (*mcp.CallToolResult, MoveFileOutput, error) {
	absFrom, err := resolveFilePath(input.Space, input.From)
	if err != nil {
		return nil, MoveFileOutput{}, errors.As(err)
	}

	absTo, err := resolveFilePath(input.Space, input.To)
	if err != nil {
		return nil, MoveFileOutput{}, errors.As(err)
	}

	if err := os.Rename(absFrom, absTo); err != nil {
		return nil, MoveFileOutput{}, errors.As(err)
	}

	log.Infof("MCP move file %s to %s", absFrom, absTo)
	return nil, MoveFileOutput{Moved: true}, nil
}

func listDirHandler(ctx context.Context, req *mcp.CallToolRequest, input ListDirInput) (*mcp.CallToolResult, ListDirOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, ListDirOutput{}, errors.As(err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, ListDirOutput{}, errors.As(err)
	}

	result := ListDirOutput{Entries: make([]FileEntry, 0, len(entries))}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result.Entries = append(result.Entries, FileEntry{
			Name:    e.Name(),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return nil, result, nil
}

func statFileHandler(ctx context.Context, req *mcp.CallToolRequest, input StatFileInput) (*mcp.CallToolResult, StatFileOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, StatFileOutput{}, errors.As(err)
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, StatFileOutput{}, errors.As(err)
	}

	return nil, StatFileOutput{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func truncateFileHandler(ctx context.Context, req *mcp.CallToolRequest, input TruncateFileInput) (*mcp.CallToolResult, TruncateFileOutput, error) {
	absPath, err := resolveFilePath(input.Space, input.Path)
	if err != nil {
		return nil, TruncateFileOutput{}, errors.As(err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, TruncateFileOutput{}, errors.As(err)
	}

	f, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, TruncateFileOutput{}, errors.As(err)
	}
	defer f.Close()

	if err := f.Truncate(input.Size); err != nil {
		return nil, TruncateFileOutput{}, errors.As(err)
	}

	log.Infof("MCP truncate file %s to %d", absPath, input.Size)
	return nil, TruncateFileOutput{Truncated: true}, nil
}

func getCapacityHandler(ctx context.Context, req *mcp.CallToolRequest, input GetCapacityInput) (*mcp.CallToolResult, CapacityInfo, error) {
	spacePath, err := resolveSpacePath(input.Space)
	if err != nil {
		return nil, CapacityInfo{}, errors.As(err)
	}

	fs := syscall.Statfs_t{}
	if err := syscall.Statfs(spacePath, &fs); err != nil {
		return nil, CapacityInfo{}, errors.As(err)
	}

	total := int64(fs.Blocks) * int64(fs.Bsize)
	avail := int64(fs.Bavail) * int64(fs.Bsize)
	free := int64(fs.Bfree) * int64(fs.Bsize)
	used := total - free

	return nil, CapacityInfo{
		Total:     total,
		Available: avail,
		Used:      used,
		Free:      free,
	}, nil
}

// ---- Initialization ----
func initMCPHandler() error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "bcstorage",
		Version: version.BuildVersion(),
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read content of a file from the bcstorage. Returns the file content as a string.",
	}, readFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write content to a file in the bcstorage. Creates parent directories if needed.",
	}, writeFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_file",
		Description: "Delete a file from the bcstorage.",
	}, deleteFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_file",
		Description: "Rename or move a file within the same space.",
	}, moveFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_directory",
		Description: "List files and directories under a given path on bcstorage.",
	}, listDirHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stat_file",
		Description: "Get file or directory metadata (size, modification time, type) from bcstorage.",
	}, statFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "truncate_file",
		Description: "Truncate a file to a specified size. Creates the file if it doesn't exist.",
	}, truncateFileHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_capacity",
		Description: "Get the storage capacity and usage of a user space on bcstorage.",
	}, getCapacityHandler)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	// Register as /mcp on the auth HTTPS server with BasicAuth middleware
	RegisterSysHandle("/mcp", func(w http.ResponseWriter, r *http.Request) error {
		if _, err := authAdmin(r); err != nil {
			writeMsg(w, http.StatusUnauthorized, errors.As(err).Code())
			return nil
		}
		handler.ServeHTTP(w, r)
		return nil
	})

	log.Infof("MCP handler registered at /mcp (on auth HTTPS port)")
	return nil
}
