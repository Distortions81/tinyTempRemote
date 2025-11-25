package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync/atomic"
)

var (
    configFilePath string
    activeConfig atomic.Value // stores Config
)

// Config mirrors the JSON configuration used by piTimeWeather.
type Config struct {
    OpenWeatherAPIKey string  `json:"openweather_api_key"`
    Latitude          float64 `json:"latitude"`
    Longitude         float64 `json:"longitude"`
    Units             string  `json:"units"`
    TimeFormat        string  `json:"time_format"`
    LocationName      string  `json:"location_name"`
    Theme             string  `json:"theme"`
    ShowSeconds       bool    `json:"show_seconds"`
    ShowLocationName  bool    `json:"show_location_name"`
    DetailHighLow     string  `json:"detail_high_low"`
    DetailWindSpeed   string  `json:"detail_wind_speed"`
    DetailWindDir     string  `json:"detail_wind_direction"`
    DetailHumidity    string  `json:"detail_humidity"`
    DetailDewPoint    string  `json:"detail_dew_point"`
    DetailPressure    string  `json:"detail_pressure"`
    DetailPrecip      string  `json:"detail_precip_chance"`
}

var defaultConfig = Config{
    OpenWeatherAPIKey: "",
    Latitude:          41.139981,
    Longitude:         -104.820246,
    Units:             "imperial",
    TimeFormat:        "12h",
    LocationName:      "Cheyenne, Wyoming, US",
    Theme:             defaultTheme,
    ShowSeconds:       false,
    ShowLocationName:  false,
    DetailHighLow:     "both",
    DetailWindSpeed:   "both",
    DetailWindDir:     "both",
    DetailHumidity:    "both",
    DetailDewPoint:    "none",
    DetailPressure:    "none",
    DetailPrecip:      "both",
}

var defaultDetailSettings = map[string]string{
    "detail_high_low":     "both",
    "detail_wind_speed":   "both",
    "detail_wind_direction": "both",
    "detail_humidity":     "both",
    "detail_dew_point":    "none",
    "detail_pressure":     "none",
    "detail_precip_chance": "both",
}

var allowedUnits = map[string]string{
    "imperial": "F",
    "metric":   "C",
}

var unitAliases = map[string]string{
    "metric":       "metric",
    "m":            "metric",
    "c":            "metric",
    "celsius":      "metric",
    "b":            "metric",
    "us":           "imperial",
    "imperial":     "imperial",
    "f":            "imperial",
    "fahrenheit":   "imperial",
    "customary":    "imperial",
    "us customary": "imperial",
    "a":            "imperial",
}

var timeFormatAliases = map[string]string{
    "12":      "12h",
    "12h":     "12h",
    "12-hour": "12h",
    "24":      "24h",
    "24h":     "24h",
    "24-hour": "24h",
}

var detailOptionAliases = map[string]string{
    "off":      "none",
    "disabled": "none",
    "disable":  "none",
    "all":      "both",
}

var detailOptionChoices = map[string]struct{}{
    "none": {},
    "today": {},
    "week": {},
    "both": {},
}

var booleanTrue = map[string]struct{}{
    "1":    {},
    "true": {},
    "on":   {},
    "yes":  {},
    "y":    {},
}

var booleanFalse = map[string]struct{}{
    "0":     {},
    "false": {},
    "off":   {},
    "no":    {},
    "n":     {},
}

func init() {
    configFilePath = determineConfigPath()
    activeConfig.Store(defaultConfig)
}

func determineConfigPath() string {
    exePath, err := os.Executable()
    if err == nil {
        dir := filepath.Dir(exePath)
        return filepath.Join(dir, "config.json")
    }
    cwd, err := os.Getwd()
    if err == nil {
        return filepath.Join(cwd, "config.json")
    }
    return "config.json"
}

func getActiveConfig() Config {
    if cfg, ok := activeConfig.Load().(Config); ok {
        return cfg
    }
    return defaultConfig
}

func setActiveConfig(cfg Config) {
    activeConfig.Store(cfg)
}

func ensureConfig() (Config, error) {
    cfg, err := loadConfig()
    if err == nil {
        setActiveConfig(cfg)
        return cfg, nil
    }
    if errors.Is(err, os.ErrNotExist) {
        if writeErr := writeConfig(defaultConfig); writeErr != nil {
            return Config{}, writeErr
        }
        setActiveConfig(defaultConfig)
        return defaultConfig, nil
    }

    raw := map[string]any{}
    if data, readErr := os.ReadFile(configFilePath); readErr == nil {
        _ = json.Unmarshal(data, &raw)
    }
    sanitized, valErr := validateConfig(raw, true)
    if valErr != nil {
        sanitized = defaultConfig
    }
    if writeErr := writeConfig(sanitized); writeErr != nil {
        return Config{}, writeErr
    }
    setActiveConfig(sanitized)
    return sanitized, nil
}

func loadConfig() (Config, error) {
    data, err := os.ReadFile(configFilePath)
    if err != nil {
        return Config{}, err
    }
    raw := map[string]any{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return Config{}, fmt.Errorf("invalid JSON: %w", err)
    }
    cfg, err := validateConfig(raw, false)
    if err != nil {
        return Config{}, err
    }
    setActiveConfig(cfg)
    return cfg, nil
}

func writeConfig(cfg Config) error {
    payload, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(configFilePath, append(payload, '\n'), 0o644)
}

func validateConfig(data map[string]any, allowEmptyKey bool) (Config, error) {
    cfg := defaultConfig

    apiKey := stringValue(data, "openweather_api_key")
    if apiKey == "" && !allowEmptyKey {
        return Config{}, errors.New("missing OpenWeather API key. Visit the dashboard Settings page to add one")
    }
    cfg.OpenWeatherAPIKey = apiKey

    lat, err := floatValue(data, "latitude")
    if err != nil {
        return Config{}, errors.New("latitude and longitude must be numbers")
    }
    cfg.Latitude = lat

    lon, err := floatValue(data, "longitude")
    if err != nil {
        return Config{}, errors.New("latitude and longitude must be numbers")
    }
    cfg.Longitude = lon

    units, err := normalizeUnit(stringValue(data, "units", cfg.Units))
    if err != nil {
        return Config{}, err
    }
    cfg.Units = units

    timeFmt, err := normalizeTimeFormat(stringValue(data, "time_format", cfg.TimeFormat))
    if err != nil {
        return Config{}, err
    }
    cfg.TimeFormat = timeFmt

    cfg.LocationName = strings.TrimSpace(stringValue(data, "location_name", cfg.LocationName))

    theme, err := normalizeTheme(stringValue(data, "theme", cfg.Theme))
    if err != nil {
        return Config{}, err
    }
    cfg.Theme = theme

    showSeconds, err := normalizeBoolean(data["show_seconds"], true)
    if err != nil {
        return Config{}, err
    }
    cfg.ShowSeconds = showSeconds

    showLocation, err := normalizeBoolean(data["show_location_name"], cfg.ShowLocationName)
    if err != nil {
        return Config{}, err
    }
    cfg.ShowLocationName = showLocation

    detailConfig := map[string]string{}
    for key, defaultValue := range defaultDetailSettings {
        rawValue := stringValue(data, key, defaultValue)
        normalized, detailErr := normalizeDetailOption(key, rawValue)
        if detailErr != nil {
            return Config{}, detailErr
        }
        detailConfig[key] = normalized
    }
    cfg.DetailHighLow = detailConfig["detail_high_low"]
    cfg.DetailWindSpeed = detailConfig["detail_wind_speed"]
    cfg.DetailWindDir = detailConfig["detail_wind_direction"]
    cfg.DetailHumidity = detailConfig["detail_humidity"]
    cfg.DetailDewPoint = detailConfig["detail_dew_point"]
    cfg.DetailPressure = detailConfig["detail_pressure"]
    cfg.DetailPrecip = detailConfig["detail_precip_chance"]

    return cfg, nil
}

func stringValue(data map[string]any, key string, fallback ...string) string {
    if data == nil {
        if len(fallback) > 0 {
            return fallback[0]
        }
        return ""
    }
    if val, ok := data[key]; ok && val != nil {
        switch typed := val.(type) {
        case string:
            return strings.TrimSpace(typed)
        case fmt.Stringer:
            return strings.TrimSpace(typed.String())
        case []byte:
            return strings.TrimSpace(string(typed))
        case int:
            return strconv.Itoa(typed)
        case int64:
            return strconv.FormatInt(typed, 10)
        case float64:
            return strconv.FormatFloat(typed, 'f', -1, 64)
        case float32:
            return strconv.FormatFloat(float64(typed), 'f', -1, 32)
        case bool:
            return strconv.FormatBool(typed)
        default:
            return strings.TrimSpace(fmt.Sprint(typed))
        }
    }
    if len(fallback) > 0 {
        return fallback[0]
    }
    return ""
}

func floatValue(data map[string]any, key string) (float64, error) {
    if data == nil {
        return 0, errors.New("missing coordinate")
    }
    if val, ok := data[key]; ok && val != nil {
        switch typed := val.(type) {
        case float64:
            return typed, nil
        case float32:
            return float64(typed), nil
        case int:
            return float64(typed), nil
        case int64:
            return float64(typed), nil
        case string:
            trimmed := strings.TrimSpace(typed)
            if trimmed == "" {
                break
            }
            return strconv.ParseFloat(trimmed, 64)
        default:
            return 0, errors.New("latitude and longitude must be numbers")
        }
    }
    return 0, errors.New("missing coordinate")
}

func normalizeTheme(value string) (string, error) {
    normalized := strings.ToLower(strings.TrimSpace(value))
    if alias := themeAliases[normalized]; alias != "" {
        normalized = alias
    }
    if _, ok := themes[normalized]; !ok {
        display := []string{}
        for _, name := range themeChoices {
            display = append(display, themeDisplayNames[name])
        }
        return "", fmt.Errorf("theme must be one of: %s", strings.Join(display, ", "))
    }
    return normalized, nil
}

func normalizeUnit(value string) (string, error) {
    normalized := strings.ToLower(strings.TrimSpace(value))
    if alias := unitAliases[normalized]; alias != "" {
        normalized = alias
    }
    if _, ok := allowedUnits[normalized]; !ok {
        return "", errors.New("units must be either Metric (°C) or US Customary (°F)")
    }
    return normalized, nil
}

func normalizeTimeFormat(value string) (string, error) {
    normalized := strings.ToLower(strings.TrimSpace(value))
    if alias := timeFormatAliases[normalized]; alias != "" {
        normalized = alias
    }
    if normalized != "12h" && normalized != "24h" {
        return "", errors.New("time format must be 12h or 24h")
    }
    return normalized, nil
}

func normalizeDetailOption(key, value string) (string, error) {
    normalized := strings.ToLower(strings.TrimSpace(value))
    if alias := detailOptionAliases[normalized]; alias != "" {
        normalized = alias
    }
    if _, ok := detailOptionChoices[normalized]; !ok {
        readable := "none, today, week, both"
        return "", fmt.Errorf("%s must be one of: %s", strings.Title(strings.ReplaceAll(key, "_", " ")), readable)
    }
    return normalized, nil
}

func normalizeBoolean(value any, fallback bool) (bool, error) {
    switch typed := value.(type) {
    case bool:
        return typed, nil
    case string:
        trimmed := strings.TrimSpace(strings.ToLower(typed))
        if trimmed == "" {
            return fallback, nil
        }
        if _, ok := booleanTrue[trimmed]; ok {
            return true, nil
        }
        if _, ok := booleanFalse[trimmed]; ok {
            return false, nil
        }
        return false, fmt.Errorf("value must be set to on or off")
    case nil:
        return fallback, nil
    default:
        return fallback, fmt.Errorf("value must be set to on or off")
    }
}
