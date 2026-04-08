package config

// DefaultGeminiTextModel は環境変数未指定時の既定テキスト生成モデルです（GEMINI_MODEL_NAME のデフォルトと一致させること）。
const DefaultGeminiTextModel = "gemini-2.5-pro"

// GeminiTextModelChoice はスラッシュコマンド等で使うテキストモデル選択肢です。
type GeminiTextModelChoice struct {
	DisplayName string
	ModelID     string
}

// GeminiTextModelChoices はサポートするテキストモデル一覧です（検証・UIで共通利用）。
func GeminiTextModelChoices() []GeminiTextModelChoice {
	return []GeminiTextModelChoice{
		{DisplayName: "Gemini 2.5 Pro", ModelID: DefaultGeminiTextModel},
		{DisplayName: "Gemini 2.0 Flash", ModelID: "gemini-2.0-flash"},
		{DisplayName: "Gemini 2.5 Flash Lite", ModelID: "gemini-2.5-flash-lite"},
	}
}

// IsSupportedGeminiTextModel は model が許可リストに含まれるかを返します。
func IsSupportedGeminiTextModel(model string) bool {
	for _, c := range GeminiTextModelChoices() {
		if c.ModelID == model {
			return true
		}
	}
	return false
}
