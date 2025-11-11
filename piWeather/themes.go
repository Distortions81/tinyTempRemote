package main

// Theme represents the palette information used by the dashboard templates.
type Theme struct {
    Label          string
    Background     string
    TextPrimary    string
    TextSecondary  string
    TextHint       string
    Accent         string
    CardBackground string
    CardText       string
    CardSubtext    string
    CardBorder     string
    CardShadow     string
}

const defaultTheme = "dark_colorful"

var themes = map[string]Theme{
    "dark_monochrome": {
        Label:          "Dark Monochrome",
        Background:     "#080808",
        TextPrimary:    "#f5f5f5",
        TextSecondary:  "#bdbdbd",
        TextHint:       "#8c8c8c",
        Accent:         "#fafafa",
        CardBackground: "#141414",
        CardText:       "#f1f1f1",
        CardSubtext:    "#c5c5c5",
        CardBorder:     "rgba(255, 255, 255, 0.06)",
        CardShadow:     "0 18px 34px rgba(0, 0, 0, 0.55)",
    },
    "light_monochrome": {
        Label:          "Light Monochrome",
        Background:     "#f8f9fa",
        TextPrimary:    "#1c1c1e",
        TextSecondary:  "#3f3f46",
        TextHint:       "#676772",
        Accent:         "#101010",
        CardBackground: "#ffffff",
        CardText:       "#202020",
        CardSubtext:    "#3f3f46",
        CardBorder:     "rgba(16, 16, 16, 0.08)",
        CardShadow:     "0 16px 32px rgba(16, 16, 16, 0.08)",
    },
    "dark_colorful": {
        Label:          "Dark Colorful",
        Background:     "#041b33",
        TextPrimary:    "#e6f4ff",
        TextSecondary:  "#a6c7dd",
        TextHint:       "#7aa2c2",
        Accent:         "#58d6ff",
        CardBackground: "#0f2c46",
        CardText:       "#e8f6ff",
        CardSubtext:    "#a8c9e1",
        CardBorder:     "rgba(88, 214, 255, 0.22)",
        CardShadow:     "0 20px 40px rgba(4, 27, 51, 0.62)",
    },
    "light_colorful": {
        Label:          "Light Colorful",
        Background:     "#eef7ff",
        TextPrimary:    "#123a52",
        TextSecondary:  "#396886",
        TextHint:       "#5c7a90",
        Accent:         "#ff6f61",
        CardBackground: "#ffffff",
        CardText:       "#123a52",
        CardSubtext:    "#4a6f87",
        CardBorder:     "rgba(18, 58, 82, 0.15)",
        CardShadow:     "0 18px 28px rgba(18, 58, 82, 0.12)",
    },
}

var themeChoices = []string{"dark_monochrome", "light_monochrome", "dark_colorful", "light_colorful"}

var themeDisplayNames = map[string]string{
    "dark_monochrome":  "Dark Monochrome",
    "light_monochrome": "Light Monochrome",
    "dark_colorful":    "Dark Colorful",
    "light_colorful":   "Light Colorful",
}

var themeAliases = map[string]string{
    "dark":         "dark_monochrome",
    "dark-mono":    "dark_monochrome",
    "dm":           "dark_monochrome",
    "light":        "light_monochrome",
    "light-mono":   "light_monochrome",
    "lm":           "light_monochrome",
    "dark-color":   "dark_colorful",
    "dark-colorful":"dark_colorful",
    "dc":           "dark_colorful",
    "light-color":  "light_colorful",
    "light-colorful":"light_colorful",
    "lc":           "light_colorful",
}
