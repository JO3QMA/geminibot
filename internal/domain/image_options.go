package domain

// ImageStyle は画像スタイルを表す定数です
type ImageStyle int

const (
	ImageStylePhotographic ImageStyle = iota
	ImageStyleAnime
	ImageStyleIllustration
	ImageStyleOilPainting
	ImageStyleWatercolor
	ImageStyleDigitalArt
	ImageStyleSketch
	ImageStyleCartoon
)

// ImageQuality は画像品質を表す定数です
type ImageQuality int

const (
	ImageQualityStandard ImageQuality = iota
	ImageQualityHigh
)

// ImageSize は画像サイズを表す定数です
type ImageSize int

const (
	ImageSize512x512 ImageSize = iota
	ImageSize1024x1024
	ImageSize1024x768
	ImageSize768x1024
)

// discordOptionData はImageStyle, ImageQuality, ImageSizeのデータを保持します
type discordOptionData struct {
	Value       string
	DisplayName string
}

// imageStyles は各ImageStyleのデータを定義します
var imageStyles = []discordOptionData{
	{"photographic", "写真風"},
	{"anime", "アニメ風"},
	{"illustration", "イラスト風"},
	{"oil_painting", "油絵風"},
	{"watercolor", "水彩画風"},
	{"digital_art", "デジタルアート風"},
	{"sketch", "スケッチ風"},
	{"cartoon", "カートゥーン風"},
}

// imageQualities は各ImageQualityのデータを定義します
var imageQualities = []discordOptionData{
	{"standard", "標準"},
	{"high", "高品質"},
}

// imageSizes は各ImageSizeのデータを定義します
var imageSizes = []discordOptionData{
	{"512x512", "512x512"},
	{"1024x1024", "1024x1024"},
	{"1024x768", "1024x768"},
	{"768x1024", "768x1024"},
}

// String はImageStyleの英語名を返します
func (s ImageStyle) String() string {
	if int(s) >= 0 && int(s) < len(imageStyles) {
		return imageStyles[s].Value
	}
	return "photographic"
}

// Japanese はImageStyleの日本語名を返します
func (s ImageStyle) DisplayName() string {
	if int(s) >= 0 && int(s) < len(imageStyles) {
		return imageStyles[s].DisplayName
	}
	return "写真風"
}

// String はImageQualityの英語名を返します
func (q ImageQuality) String() string {
	if int(q) >= 0 && int(q) < len(imageQualities) {
		return imageQualities[q].Value
	}
	return "standard"
}

// Japanese はImageQualityの日本語名を返します
func (q ImageQuality) DisplayName() string {
	if int(q) >= 0 && int(q) < len(imageQualities) {
		return imageQualities[q].DisplayName
	}
	return "標準"
}

// String はImageSizeの英語名を返します
func (s ImageSize) String() string {
	if int(s) >= 0 && int(s) < len(imageSizes) {
		return imageSizes[s].Value
	}
	return "512x512"
}

// Japanese はImageSizeの日本語名を返します
func (s ImageSize) DisplayName() string {
	if int(s) >= 0 && int(s) < len(imageSizes) {
		return imageSizes[s].DisplayName
	}
	return "512x512"
}

// AllImageStyles はすべてのImageStyleを返します
func AllImageStyles() []ImageStyle {
	return []ImageStyle{
		ImageStylePhotographic,
		ImageStyleAnime,
		ImageStyleIllustration,
		ImageStyleOilPainting,
		ImageStyleWatercolor,
		ImageStyleDigitalArt,
		ImageStyleSketch,
		ImageStyleCartoon,
	}
}

// AllImageQualities はすべてのImageQualityを返します
func AllImageQualities() []ImageQuality {
	return []ImageQuality{
		ImageQualityStandard,
		ImageQualityHigh,
	}
}

// AllImageSizes はすべてのImageSizeを返します
func AllImageSizes() []ImageSize {
	return []ImageSize{
		ImageSize512x512,
		ImageSize1024x1024,
		ImageSize1024x768,
		ImageSize768x1024,
	}
}
