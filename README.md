# Go Downloader

一个使用 Go 语言编写的多线程下载器，用于高效地下载大文件。

## 特性

- 多线程并发下载，提高下载速度
- 实时显示下载进度、速度和预计剩余时间
- 支持断点续传和分片下载
- 自动检测服务器是否支持范围请求
- 自动回退到单线程下载（当服务器不支持范围请求时）
- 失败重试机制
- 简单易用的命令行界面

## 安装

### 从源码安装

```bash
git clone https://github.com/yourusername/go-downloader.git
cd go-downloader
go build -o godownloader ./cmd/downloader
```

### 使用 Go Install

```bash
go install github.com/yourusername/go-downloader/cmd/downloader@latest
```

## 使用方法

### 命令行

```bash
# 基本用法
godownloader -url https://example.com/largefile.zip

# 指定输出文件
godownloader -url https://example.com/largefile.zip -output myfile.zip

# 指定线程数
godownloader -url https://example.com/largefile.zip -threads 8

# 静默模式
godownloader -url https://example.com/largefile.zip -quiet

# 查看帮助
godownloader -help
```

### 作为库使用

```go
package main

import (
    "fmt"
    "os"

    "github.com/yourusername/go-downloader/pkg/downloader"
)

func main() {
    // 基本用法
    dl := downloader.New("https://example.com/largefile.zip", "output.zip", 4)
    err := dl.Download()
    if err != nil {
        fmt.Printf("下载失败: %v\n", err)
        os.Exit(1)
    }

    // 使用自定义选项
    options := downloader.Options{
        OutputPath: "custom-output.zip",
        NumThreads: 8,
        MaxRetries: 5,
        Verbose:    true,
    }
    dl = downloader.WithOptions("https://example.com/largefile.zip", options)
    err = dl.Download()
    if err != nil {
        fmt.Printf("下载失败: %v\n", err)
        os.Exit(1)
    }
}
```

## 命令行参数

| 参数       | 描述                     | 默认值                |
| ---------- | ------------------------ | --------------------- |
| `-url`     | 要下载的 URL             | -                     |
| `-output`  | 输出文件路径             | 从 URL 中提取的文件名 |
| `-threads` | 下载线程数               | CPU 核心数            |
| `-retries` | 失败重试次数             | 3                     |
| `-quiet`   | 静默模式，只显示错误信息 | false                 |
| `-version` | 显示版本信息             | false                 |

## 工作原理

1. 发送 HEAD 请求获取文件大小和范围支持信息
2. 如果服务器支持范围请求，则将文件分割为多个块
3. 为每个块创建一个 worker 并发下载
4. 实时跟踪和显示下载进度
5. 合并所有块为最终文件
6. 清理临时文件

## 许可证

MIT
