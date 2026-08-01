# 浮水印影片匯出編碼優化指南 (Encoding Optimization Guide)

## 一、 問題背景與原因分析

在 [nvr_core/service/export.go](file:///C:/workspace/rm_vms/nvr_core/service/export.go) 的單次 Pass 浮水印壓制中，產出的影片檔案大小為原始檔案的 2 倍以上。主要原因如下：

### 1. 無浮水印 vs. 有浮水印的機制差異
- **原始無浮水印 (`-c copy`)**：直接封裝攝影機原生的 H.264/H.265 碼率（通常為 1.5 ~ 2.0 Mbps），不進行解碼與重編。
- **加浮水印 (Transcode)**：必須逐幀解碼並透過 `libx264` 進行二次壓縮。

### 2. Preset `superfast` 與 CRF `23` 的連動效應 (最主因)
- **CRF (Constant Rate Factor)** 是 x264 的畫質控制指標（預設 23）。
- **`superfast` 預設集** 為了省 CPU 算力，關閉了大量運動估計與區塊預測演算法。
- 在相同的 CRF 23 設定下，`superfast` 產出的檔案位元率會比 `medium` 大 1.5 ~ 2.5 倍，導致最終檔案比攝影機原始檔大 2 倍以上。

---

## 二、 優化方案與建議參數 (不犧牲效率與品質)

在維持快速編碼（不增加 CPU 負載）的前提下，推薦以下兩種優化方案：

### 方案 1：維持 `superfast` + 調整 CRF 至 `26` ~ `27` (首選推薦)

- **FFmpeg 參數**：
  ```bash
  -c:v libx264 -preset superfast -crf 27
  ```
- **特點**：
  - **編碼速度 100% 維持原狀**（完全不增加 CPU 消耗）。
  - **檔案大小大幅縮減 40% ~ 60%**（恢復至接近原始錄影檔大小）。
  - **視覺品質良好**：對於 1080p 監控影片，肉眼無明顯降質感受。

### 方案 2：調整為 `veryfast` + CRF `25`

- **FFmpeg 參數**：
  ```bash
  -c:v libx264 -preset veryfast -crf 25
  ```
- **特點**：
  - 啟動稍佳的運動補償算法，畫面細節保留更佳。
  - CPU 負擔僅微幅增加約 5%~10%。
  - 檔案大小縮減約 50%。

### 補充安全防護：限制最大碼率 (Max Bitrate Cap)
避免極端畫面（如樹葉搖晃、雨天雜訊）導致 CRF 分配過高位元率，可搭配限制最大碼率：
```bash
-c:v libx264 -preset superfast -crf 27 -maxrate 4M -bufsize 8M
```

---

## 三、 方案對比總覽

| 方案 | FFmpeg 編碼參數 | 預估檔案大小 | CPU 編碼速度 | 視覺品質 |
|------|-----------------|--------------|--------------|----------|
| **現行實作** | `-preset superfast -crf 23` | **100% (約原檔 2x)** | 100% (極快) | 高 |
| **優化方案 1 (推薦)** | `-preset superfast -crf 27` | **~45% (接近原檔)** | **100% (極快)** | 良好 (無明顯差異) |
| **優化方案 2** | `-preset veryfast -crf 25` | **~50% (接近原檔)** | ~90% (快) | 高 |

---

## 四、 未來執行階段修改位置參考

欲套用此優化時，只需修改 [nvr_core/service/export.go](file:///C:/workspace/rm_vms/nvr_core/service/export.go#L240) 中的 `executeFFmpeg` 方法：

```go
// 尋找此行：
ffmpegArgs = append(ffmpegArgs, "-c:v", "libx264", "-preset", "superfast", "-crf", "23")

// 替換為 (優化方案 1)：
ffmpegArgs = append(ffmpegArgs, "-c:v", "libx264", "-preset", "superfast", "-crf", "27")
```
