# envault

> Lightweight .env secret manager with per-project encryption and shell integration.

---

## Installation

```bash
go install github.com/yourname/envault@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourname/envault/releases).

---

## Usage

Initialize a vault for your project:

```bash
envault init
```

Add and retrieve secrets:

```bash
# Store a secret
envault set DATABASE_URL "postgres://user:pass@localhost/mydb"

# Retrieve a secret
envault get DATABASE_URL

# Load all secrets into your shell session
eval $(envault export)
```

Run a command with secrets injected into the environment:

```bash
envault run -- go run main.go
```

Secrets are encrypted per-project using AES-256-GCM and stored in a local `.envault` file. Each project derives its own encryption key, keeping secrets isolated across repositories.

---

## Shell Integration

Add the following to your `.bashrc` or `.zshrc` to auto-load secrets when entering a project directory:

```bash
source <(envault hook)
```

---

## License

[MIT](LICENSE) © 2024 yourname