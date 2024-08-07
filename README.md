# AirClipboard

English | [中文](./README_zh.md)

---

[AirClipboard](https://airclipbd.com) combines the convenience of Snapdrop with the functionality of an online clipboard, enabling seamless file sharing and clipboard management between devices.

## Features

### 1. Snapdrop

[Snapdrop](https://github.com/RobinLinus/snapdrop) is a browser-based local file-sharing service inspired by Apple's AirDrop. It allows devices on the same network to automatically discover each other and supports peer-to-peer file transfer between devices.

**Features:**
- **Automatic Device Discovery:** Devices on the same network automatically discover each other without the need for manual configuration.
- **Peer-to-Peer File Transfer:** Direct file transfers between devices ensure fast and secure communication.

### 2. Online Clipboard

The online clipboard provides a simple and convenient way to manage clipboard content.

**Features:**
- **Clipboard Space:** Create and access clipboard spaces using `/${board_name}`.
- **Public Read/Write Access:** Open access for easy reading and writing of content.
- **Content Support:** Directly paste clipboard content, supporting text, images, and various file formats.
- **Content Limitations:** Maximum paste size is 20MB, and each clipboard space temporarily stores the latest 20 entries.
- **Caching Support:** Supports both local memory and Redis cache to ensure efficient clipboard data management.

## Installation

To install and start using AirClipboard, follow these steps:

### Prerequisites

- Ensure [Go](https://golang.org/dl/) 1.19 or later is installed.
- Optional: Install [Docker](https://www.docker.com/) for containerized deployment.

### Steps

1. **Clone the Project Repository**

   ```bash
   git clone https://github.com/r0n9/airclipboard.git
   cd airclipboard
   ```

2. **Install Dependencies**

   ```bash
   go mod tidy
   ```

3. **Build the Project**

   ```bash
   go build -o airclipboard
   ```

4. **Run the Application**

   You can configure the program using command-line arguments when running:

   ```bash
   ./airclipboard --cache-type=redis --redis-addr=localhost:6379 --redis-password=yourpassword --redis-db=0
   ```

    - **Available Parameters:**
        - `--cache-type`: Type of cache, can be `memory` or `redis`. Defaults to `memory`.
        - `--redis-addr`: Address of the Redis server, defaults to `localhost:6379`.
        - `--redis-password`: Password for the Redis server (if needed), defaults to `******`.
        - `--redis-db`: Redis database number, defaults to `0`.

5. **Alternatively, Start with Docker**

    - Build the Docker image:

      ```bash
      docker build -t airclipboard .
      ```

    - Run the Docker container:

      ```bash
      docker run -p 18128:18128 airclipboard --cache-type=redis --redis-addr=localhost:6379 --redis-password=yourpassword --redis-db=0
      ```
6. **visit `http://your-host-ip:18128` and enjoy**

## Contributing

We welcome contributions from the community. If you wish to contribute code, please Fork the repository and submit a Pull Request. For major changes, please open an Issue first to discuss your proposals.

## License

AirClipboard is open-source and available under the [GNU License](LICENSE).