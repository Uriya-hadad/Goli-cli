# Goli CLI

**Goli** is a powerful command-line interface designed to simplify the management of applications and services in **Cloud Foundry** environments.

## Why Goli?

Goli is more than just a Cloud Foundry CLI wrapper it's a productivity-focused tool with built-in intelligence and developer experience in mind:

### Automatic Updates

No more manual downloads or version checks.
Once Goli is installed, it keeps itself up-to-date.
Each time you exit the CLI, it checks for new releases and seamlessly applies them in the background — ensuring you always run the latest and most secure version.

### Global and Team-Specific Features

Goli includes a dual-layer command model:

- **Global Features**: Available to all users across environments. These are the core tools for managing applications, services, and environments.
- **Team Features**: Role-specific capabilities that unlock advanced functionality (e.g., DB query, run a specific process) depending on your organizational role. Goli securely detects your permissions and reveals only the features relevant to you.

### High Performance with Caching

Behind the scenes, Goli intelligently caches metadata like app/service lists, environment details, and credential info. This drastically reduces repeated network calls to Cloud Foundry APIs, making command executions feel instant.

Whether you're querying services or navigating instances interactively, Goli is optimized for speed and efficiency.

## Overview

Goli provides a rich set of commands to:

- Manage Cloud Foundry applications
- Modify environment variables
- Connect to PostgreSQL and Redis via SSH tunnels
- View logs and app status
- Automate workflows with interactive support

### Examples
![goli-query](https://github.com/user-attachments/assets/a2cbca8d-3351-4ca1-9fd0-5e49bd2b05b0)
![goli-open-DB](https://github.com/user-attachments/assets/3d33cb67-58f8-4c84-93fe-5f589de5e27a)
![goli-gen-key-interactive](https://github.com/user-attachments/assets/ab4ac02e-42b3-42f9-a260-fc8c9a4db98e)
![goli-gen-key-new](https://github.com/user-attachments/assets/32cd5f8c-443c-4d47-98d2-8c2e672b8e90)

## Documentation

All usage instructions, command references, and advanced options are available in the official documentation:

**[Goli docs](https://goli-cli.cfapps.eu12.hana.ondemand.com/goli/)**

This includes:

- CLI command reference
- Environment variable handling
- Interactive vs. scripted modes
- Team-specific tools
- Auto-completion setup
