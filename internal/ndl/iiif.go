package ndl

import "fmt"

// NDLのIIIF Image APIは `/iiif/<PID>/<region>/<size>/<rotation>/<quality>` 系。
// ここでは `pct:x,y,w,h` と ページ画像の基本パスを渡してURLを合成する素朴版。

func BuildIIIFCropURL(iiifImageBase, bboxPct string) string {
	// 例: iiifImageBase = https://dl.ndl.go.jp/api/iiif/2558316/28/full/800,/0/default.jpg のようなベースを想定
	// 実運用では manifest から page毎の canvas → image service を辿って base を得る。
	if iiifImageBase == "" || bboxPct == "" {
		return ""
	}
	// sizeはfullの代わりに `,` 指定で幅自動などにしても良い
	return fmt.Sprintf("%s/%s/800,/0/default.jpg", iiifImageBase, bboxPct)
}
