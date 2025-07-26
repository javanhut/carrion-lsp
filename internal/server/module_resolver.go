package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModuleResolver handles import path resolution following Carrion's module system
type ModuleResolver struct {
	WorkspaceRoot   string   // Root directory of the workspace
	CarrionPath     string   // Path to Carrion installation (optional)
	UserPackagesDir string   // ~/.carrion/packages/
	GlobalLibDir    string   // /usr/local/share/carrion/lib/
	BuiltinModules  []string // List of built-in module names
}

// ModuleInfo contains information about a resolved module
type ModuleInfo struct {
	Name       string // Module name as imported
	FilePath   string // Absolute path to the module file
	IsBuiltin  bool   // Whether this is a built-in module
	IsStdLib   bool   // Whether this is from Munin standard library
	PackageDir string // Directory containing the module (for relative imports within package)
}

// NewModuleResolver creates a new module resolver
func NewModuleResolver(workspaceRoot, carrionPath string) *ModuleResolver {
	homeDir, _ := os.UserHomeDir()
	userPackagesDir := filepath.Join(homeDir, ".carrion", "packages")

	return &ModuleResolver{
		WorkspaceRoot:   workspaceRoot,
		CarrionPath:     carrionPath,
		UserPackagesDir: userPackagesDir,
		GlobalLibDir:    "/usr/local/share/carrion/lib",
		BuiltinModules:  getBuiltinModules(),
	}
}

// ResolveImport resolves an import statement to an actual file path
// Follows Carrion's import resolution order:
// 1. Local files (current directory)
// 2. Project packages (./carrion_modules/)
// 3. User packages (~/.carrion/packages/)
// 4. Global packages (/usr/local/share/carrion/lib/)
// 5. Standard library (Munin)
func (mr *ModuleResolver) ResolveImport(moduleName, currentFile string) (*ModuleInfo, error) {
	// Get the directory of the current file
	currentDir := filepath.Dir(currentFile)

	// Convert URI to file path if needed
	if strings.HasPrefix(currentFile, "file://") {
		currentFile = strings.TrimPrefix(currentFile, "file://")
		currentDir = filepath.Dir(currentFile)
	}

	// 1. Check if it's a built-in module
	if mr.isBuiltinModule(moduleName) {
		return &ModuleInfo{
			Name:      moduleName,
			FilePath:  "", // Built-ins don't have file paths
			IsBuiltin: true,
			IsStdLib:  false,
		}, nil
	}

	// 2. Local files (current directory)
	if modulePath := mr.checkLocalFile(currentDir, moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   false,
			PackageDir: currentDir,
		}, nil
	}

	// 3. Project packages (./carrion_modules/)
	if modulePath := mr.checkProjectPackages(currentDir, moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   false,
			PackageDir: filepath.Dir(modulePath),
		}, nil
	}

	// 4. User packages (~/.carrion/packages/)
	if modulePath := mr.checkUserPackages(moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   false,
			PackageDir: filepath.Dir(modulePath),
		}, nil
	}

	// 5. Global packages (/usr/local/share/carrion/lib/)
	if modulePath := mr.checkGlobalPackages(moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   false,
			PackageDir: filepath.Dir(modulePath),
		}, nil
	}

	// 6. Standard library (Munin)
	if modulePath := mr.checkStandardLibrary(moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   true,
			PackageDir: filepath.Dir(modulePath),
		}, nil
	}

	return nil, fmt.Errorf("module '%s' not found", moduleName)
}

// checkLocalFile looks for the module in the current directory
func (mr *ModuleResolver) checkLocalFile(currentDir, moduleName string) string {
	// Try different file patterns
	patterns := []string{
		fmt.Sprintf("%s.crl", moduleName),
		fmt.Sprintf("%s.carrion", moduleName), // Legacy support
		filepath.Join(moduleName, "init.crl"),
		filepath.Join(moduleName, "__init__.crl"),
	}

	for _, pattern := range patterns {
		fullPath := filepath.Join(currentDir, pattern)
		if mr.fileExists(fullPath) {
			return fullPath
		}
	}

	return ""
}

// checkProjectPackages looks in ./carrion_modules/
func (mr *ModuleResolver) checkProjectPackages(currentDir, moduleName string) string {
	// Walk up the directory tree to find carrion_modules
	dir := currentDir
	for dir != "/" && dir != "." {
		carrionModulesDir := filepath.Join(dir, "carrion_modules")
		if mr.dirExists(carrionModulesDir) {
			if modulePath := mr.checkPackageDir(carrionModulesDir, moduleName); modulePath != "" {
				return modulePath
			}
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

// checkUserPackages looks in ~/.carrion/packages/
func (mr *ModuleResolver) checkUserPackages(moduleName string) string {
	if mr.dirExists(mr.UserPackagesDir) {
		return mr.checkPackageDir(mr.UserPackagesDir, moduleName)
	}
	return ""
}

// checkGlobalPackages looks in /usr/local/share/carrion/lib/
func (mr *ModuleResolver) checkGlobalPackages(moduleName string) string {
	if mr.dirExists(mr.GlobalLibDir) {
		return mr.checkPackageDir(mr.GlobalLibDir, moduleName)
	}
	return ""
}

// checkStandardLibrary looks for Munin standard library modules
func (mr *ModuleResolver) checkStandardLibrary(moduleName string) string {
	// If we have a Carrion installation path, check its standard library
	if mr.CarrionPath != "" {
		stdlibPaths := []string{
			filepath.Join(mr.CarrionPath, "src", "munin", fmt.Sprintf("%s.crl", moduleName)),
			filepath.Join(mr.CarrionPath, "lib", fmt.Sprintf("%s.crl", moduleName)),
		}

		for _, path := range stdlibPaths {
			if mr.fileExists(path) {
				return path
			}
		}
	}

	// Check common standard library locations
	commonPaths := []string{
		fmt.Sprintf("/usr/local/share/carrion/munin/%s.crl", moduleName),
		fmt.Sprintf("/usr/share/carrion/munin/%s.crl", moduleName),
	}

	for _, path := range commonPaths {
		if mr.fileExists(path) {
			return path
		}
	}

	return ""
}

// checkPackageDir looks for a module within a package directory
func (mr *ModuleResolver) checkPackageDir(packageDir, moduleName string) string {
	patterns := []string{
		filepath.Join(packageDir, fmt.Sprintf("%s.crl", moduleName)),
		filepath.Join(packageDir, moduleName, "init.crl"),
		filepath.Join(packageDir, moduleName, "__init__.crl"),
		filepath.Join(packageDir, moduleName, fmt.Sprintf("%s.crl", moduleName)),
	}

	for _, pattern := range patterns {
		if mr.fileExists(pattern) {
			return pattern
		}
	}

	return ""
}

// isBuiltinModule checks if a module is a built-in module
func (mr *ModuleResolver) isBuiltinModule(moduleName string) bool {
	for _, builtin := range mr.BuiltinModules {
		if builtin == moduleName {
			return true
		}
	}
	return false
}

// fileExists checks if a file exists
func (mr *ModuleResolver) fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists
func (mr *ModuleResolver) dirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// getBuiltinModules returns the list of built-in modules
func getBuiltinModules() []string {
	return []string{
		// Built-in modules that don't have file representations
		"file",
		"http",
		"os",
		"sockets",
		"time",
		"math",
		"json",
		"sys",
		"io",
	}
}

// GetWorkspaceFiles returns all Carrion files in the workspace
func (mr *ModuleResolver) GetWorkspaceFiles() ([]string, error) {
	var carrionFiles []string

	err := filepath.Walk(mr.WorkspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking, ignore errors
		}

		// Skip hidden directories and node_modules-like directories
		if info.IsDir() {
			name := filepath.Base(path)
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "carrion_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a Carrion file
		if strings.HasSuffix(path, ".crl") || strings.HasSuffix(path, ".carrion") {
			carrionFiles = append(carrionFiles, path)
		}

		return nil
	})

	return carrionFiles, err
}

// ResolveRelativeImport resolves imports relative to a specific package
func (mr *ModuleResolver) ResolveRelativeImport(moduleName, packageDir string) (*ModuleInfo, error) {
	if modulePath := mr.checkLocalFile(packageDir, moduleName); modulePath != "" {
		return &ModuleInfo{
			Name:       moduleName,
			FilePath:   modulePath,
			IsBuiltin:  false,
			IsStdLib:   false,
			PackageDir: packageDir,
		}, nil
	}

	return nil, fmt.Errorf("relative module '%s' not found in package '%s'", moduleName, packageDir)
}
