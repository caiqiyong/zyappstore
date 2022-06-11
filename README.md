# 掌阅应用商店

### 环境

1. windows 7 x64
2. go 1.18.1

### 测试环境

掌阅smart x 阅读器

### 编译

```json
go build
```

### 启动

```json
zyappstore.exe -port [port]
```

### 商店维护

在applist.json 文件中添加应用配置即可。满足以下约束

- id 整型 全局唯一
- appName 字符串 全局唯一
- appurl 内置文件服务器URL 是/app/, 使用内部文件服务器要把apk打包放在app目录下。

```json
{
        "id": 7,
        "name": "CXFileManager",
        "icon": "",
        "appVersion": "V1.8.7",
        "appSize": "5.43MB",
        "appName": "com.cxinventor.file.explorer",
        "appUrl": "http://10.1.2.223:9998/app/com.cxinventor.file.explorer_1.8.7.zip",
        "appDesc": "",
        "explain": "",
        "categoryId": 0
    }
```