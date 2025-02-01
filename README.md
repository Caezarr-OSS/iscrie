# Iscrie Documentation

## Overview

![Iscrie-Logo](./assets/img/iscrie.png)

**Iscrie** is a CLI tool designed to handle the automated upload of files to a Nexus repository. It supports both **RAW** and **Maven2** repository types and is highly configurable to suit a variety of deployment scenarios. The tool ensures ease of use and provides detailed logs for file uploads.

---

## Features

1. **Support for Nexus Repository**:
   - Upload files to **RAW** and **Maven2** repository types.
   - Verify the existence of repositories before processing.

2. **Configuration via `TOML`**:
   - Fully customizable settings for repository URL, authentication, proxy, and retries.

3. **File Processing**:
   - Upload files directly from a directory (`root_path`).
   - Automatically detect file types based on the repository type.

4. **Proxy Support**:
   - Configure a proxy server for HTTP requests.

5. **Error Handling and Retries**:
   - Built-in retries with configurable timeout and retry limits.

6. **Detailed Logging**:
   - Logs progress and errors for all uploads.

---

## Configuration

The configuration for **Iscrie** is stored in a `TOML` file (default: `iscrie.toml`). Below is an explanation of the configuration fields:

### General Settings

```toml
[general]
root_path = "./files"       # The root directory containing files to upload.
log_path = "./logs"         # Path to save logs.
log_level = "info"          # Log verbosity: "debug", "info", "error".
batch_size = 1              # Batch size for uploads (not currently used).           # If true, simulate actions without uploading files.
```

### Nexus Settings

```toml
[nexus]
url = "http://localhost:8081"   # Base URL of the Nexus repository.
repository = "my-repo"          # Name of the repository.
repository_type = "maven2"      # Repository type: "raw" or "maven2".
force_replace = false           # If true, overwrite existing files.
```

### Retry Settings

```toml
[retry]
retry_attempts = 3   # Number of retries for failed uploads.
timeout = 10         # Timeout in seconds for each retry.
```

### Proxy Settings

```toml
[proxy]
enabled = false        # Enable or disable proxy.
host = "proxy.example.com" # Proxy host.
port = 8080            # Proxy port.
username = ""          # Proxy username (optional).
password = ""          # Proxy password (optional).
```

### Authentication Settings

```toml
[auth]
type = "basic"          # Authentication type: "basic", "bearer", "header".
user_token = "user"     # Username for basic auth.
pass_token = "password" # Password for basic auth.
access_token = ""       # Access token for bearer auth.
header_name = ""        # Header name for custom header auth.
header_value = ""       # Header value for custom header auth.
```

---

## Usage

### 1. Run the Binary

Build the binary:

```bash
go build -o iscrie ./cmd
```

Run the binary:

```bash
./iscrie --config="iscrie.toml"
```

### 2. File Processing

**Iscrie** automatically detects the repository type (`raw` or `maven2`) and processes files accordingly.

#### RAW Repository:
- Files are uploaded "as-is."
- The directory structure under `root_path` is mirrored in the repository.

**Example**:
- File: `./files/raw/example.txt`
- Repository Path: `/raw/example.txt`

#### Maven2 Repository:
- Files must follow the Maven2 directory structure:
  - GroupID: Derived from folder hierarchy.
  - ArtifactID: Derived from the folder name above the version folder.
  - Version: Folder name containing the artifact files.

**Example Structure**:
```plaintext
files/maven2/
└── com/
    └── example/
        └── mylib/
            └── 1.0.0/
                ├── mylib-1.0.0.jar
                ├── mylib-1.0.0.pom
```

- GroupID: `com.example`
- ArtifactID: `mylib`
- Version: `1.0.0`

---

## HTTP Client

The **HTTPClient** and **HTTPClientAdapter** in `http_client.go` manage all HTTP interactions, including authentication and proxy configurations.

### Features:
1. **Authentication**:
   - Basic Auth: Username and password.
   - Bearer Token: Token-based authentication.
   - Header Auth: Custom headers for advanced use cases.

   

2. **Proxy Support**:
   - Configurable proxy host, port, username, and password.

3. **Retry Mechanism**:
   - Automatically retries failed requests with a configurable timeout.

---

## Additional Commands

### Integration Runner

The `integration_test_runner` provides an interactive CLI for performing various operations:

```bash
go run ./cmd/integration_test_runner
```

**Options:**
1. Generate Test Data: Creates sample files for testing.
2. Setup Nexus Repository: Initializes the Nexus repository.
3. Upload Files to Nexus: Uploads files from `root_path`.
4. Validate Nexus Upload: Verifies that all files exist in the repository.
5. Cleanup Test Data: Deletes test data.

---

## Logs and Debugging

Logs are stored in the path specified by `log_path` in the configuration file. Adjust the log level (`log_level`) for more or less verbosity.

---

## Example Configuration

```toml
[general]
root_path = "./test/files"
log_path = "./test/logs"
log_level = "debug"

[nexus]
url = "http://localhost:8081"
repository = "my-repo"
repository_type = "maven2"
force_replace = true

[retry]
retry_attempts = 5
timeout = 15

[proxy]
enabled = true
host = "proxy.example.com"
port = 8080
username = "proxyuser"
password = "proxypass"

[auth]
type = "basic"
user_token = "admin"
pass_token = "admin123"
```

---

## Conclusion

**Iscrie** simplifies the process of uploading files to Nexus repositories. With robust configuration options, support for Maven2 and RAW repositories, and features like retry mechanisms, it is a powerful tool for managing large-scale uploads. Adjust the configuration file to match your environment, and you're ready to go!
