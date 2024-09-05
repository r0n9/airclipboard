# 隔空剪贴板 AirClipboard

[English](./README.md) | 中文

---

[AirClipboard](https://airclipbd.com) 将 Snapdrop 的便捷性与在线剪贴板的功能相结合，实现设备之间的无缝文件共享和剪贴板管理。

## 功能说明

### 1. Snapdrop

[Snapdrop](https://github.com/RobinLinus/snapdrop) 是一个浏览器中的本地文件共享服务，灵感来自苹果的 AirDrop。它允许同一网络内的设备自动发现彼此，并支持设备间的点对点文件传输。

**功能特点：**
- **自动设备发现：** 同一网络内的设备自动相互发现，无需手动配置。
- **点对点文件传输：** 设备间直接文件传输，确保快速且安全的通信。

### 2. 在线剪贴板

在线剪贴板提供了一种简单便捷的方式来管理剪贴板内容。

**功能特点：**
- **剪贴板空间：** 使用 `/${board_name}` 创建并访问剪贴板空间。
- **公开读写访问：** 开放式访问，便于内容的读取和写入。
- **内容支持：** 直接粘贴剪贴板内容，支持文字、图片及各种文件格式。
- **内容限制：** 粘贴内容最大限制为 20MB，每个剪贴板空间暂存最新的 20 条记录。
- **缓存支持：** 支持本地内存和 Redis 缓存，确保高效的剪贴板数据管理。

## 安装

要安装并开始使用 AirClipboard，请按照以下步骤操作：

### 前提条件

- 确保已安装 [Go](https://golang.org/dl/) 1.19 或更高版本。
- 可选：安装 [Docker](https://www.docker.com/) 以便使用容器化部署。

### 步骤

1. **克隆项目仓库**

   ```bash
   git clone https://github.com/r0n9/airclipboard.git
   cd airclipboard
   ```

2. **安装依赖**

   ```bash
   go mod tidy
   ```

3. **构建项目**

   ```bash
   go build -o airclipboard
   ```

4. **运行应用程序**

   运行程序时可以使用命令行参数进行配置：

   ```bash
   ./airclipboard --cache-type=redis --redis-addr=localhost:6379 --redis-password=yourpassword --redis-db=0
   ```

    - **可用参数：**
        - `--cache-type`：缓存类型，可以是 `memory` 或 `redis`。默认为 `memory`。
        - `--redis-addr`：Redis 服务器的地址，默认为 `localhost:6379`。
        - `--redis-password`：Redis 服务器的密码（如果需要），默认为 `******`。
        - `--redis-db`：Redis 数据库编号，默认为 `0`。

5. **或使用 Docker 启动**

    - 运行 Docker 容器：

      ```bash
      # start with memory cache
      docker run -p 18128:18128 r0n9/airclipboard
      
      # start with redis cache
      docker run -p 18128:18128 r0n9/airclipboard --cache-type=redis --redis-addr=localhost:6379 --redis-password=yourpassword --redis-db=0
      ```

6. **即可访问 `http://your-host-ip:18128`**

## 贡献

我们欢迎社区的贡献。如果您希望贡献代码，请先 Fork 仓库并提交 Pull Request。对于重大更改，请先打开 Issue 以讨论您的建议。

## 许可证

隔空剪贴板 AirClipboard 是开源的，并根据 [GNU 许可证](LICENSE) 提供。
