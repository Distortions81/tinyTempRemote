package main

import (
    "errors"
    "fmt"
    "html/template"
    "log"
    "math"
    "net/http"
    "strconv"
    "strings"
    "time"
)

var (
    dashboardTemplate = template.Must(template.New("dashboard").Parse(dashboardTemplateHTML))
    settingsTemplate  = template.Must(template.New("settings").Parse(settingsTemplateHTML))
    themeOptions      = buildThemeOptions()
)

type Option struct {
    Value string
    Label string
}

type DetailField struct {
    Name  string
    Label string
}

type FormValues map[string]string

func (f FormValues) Get(key string) string {
    if f == nil {
        return ""
    }
    return f[key]
}

func (f FormValues) GetOrDefault(key, fallback string) string {
    if f == nil {
        return fallback
    }
    if value := strings.TrimSpace(f[key]); value != "" {
        return value
    }
    return fallback
}

type DashboardView struct {
    Theme        Theme
    SettingsURL  string
    City         string
    Temp         string
    Desc         string
    Unit         string
    Forecast     []ForecastCard
    ForecastError string
    TimeFormat   string
    ShowSeconds  bool
    Error        string
    Hint         string
    TodayDetails []DetailItem
}

type SettingsView struct {
    FormValues           FormValues
    UnitOptions          []Option
    TimeFormatOptions    []Option
    ThemeOptions         []Option
    DetailFields         []DetailField
    DetailOptions        []Option
    Status               string
    Error                string
    DefaultDetailSettings map[string]string
    DefaultShowLocation  string
    DefaultTheme         string
    DashboardURL         string
}

var detailFields = []DetailField{
    {Name: "detail_high_low", Label: "High / low temperature"},
    {Name: "detail_wind_speed", Label: "Wind speed"},
    {Name: "detail_wind_direction", Label: "Wind direction"},
    {Name: "detail_humidity", Label: "Humidity"},
    {Name: "detail_dew_point", Label: "Dew point"},
    {Name: "detail_pressure", Label: "Air pressure"},
    {Name: "detail_precip_chance", Label: "Precipitation chance"},
}

var detailOptions = []Option{
    {Value: "none", Label: "None"},
    {Value: "today", Label: "Today"},
    {Value: "week", Label: "Week"},
    {Value: "both", Label: "Today + Week"},
}

var unitOptions = []Option{
    {Value: "imperial", Label: "US Customary (°F)"},
    {Value: "metric", Label: "Metric (°C)"},
}

var timeFormatOptions = []Option{
    {Value: "12h", Label: "12-hour"},
    {Value: "24h", Label: "24-hour"},
}

func buildThemeOptions() []Option {
    opts := make([]Option, 0, len(themeChoices))
    for _, key := range themeChoices {
        opts = append(opts, Option{Value: key, Label: themeDisplayNames[key]})
    }
    return opts
}

func defaultDetailValues(cfg Config) FormValues {
    values := FormValues{
        "openweather_api_key": cfg.OpenWeatherAPIKey,
        "location_name":       cfg.LocationName,
        "latitude":            strconv.FormatFloat(cfg.Latitude, 'f', 6, 64),
        "longitude":           strconv.FormatFloat(cfg.Longitude, 'f', 6, 64),
        "units":               cfg.Units,
        "time_format":         cfg.TimeFormat,
        "theme":               cfg.Theme,
        "show_seconds":        onOff(cfg.ShowSeconds),
        "show_location_name":  onOff(cfg.ShowLocationName),
        "detail_high_low":     cfg.DetailHighLow,
        "detail_wind_speed":   cfg.DetailWindSpeed,
        "detail_wind_direction": cfg.DetailWindDir,
        "detail_humidity":     cfg.DetailHumidity,
        "detail_dew_point":    cfg.DetailDewPoint,
        "detail_pressure":     cfg.DetailPressure,
        "detail_precip_chance": cfg.DetailPrecip,
    }
    return values
}

func formValuesFromRequest(r *http.Request) FormValues {
    values := FormValues{
        "openweather_api_key": strings.TrimSpace(r.FormValue("openweather_api_key")),
        "location_name":       strings.TrimSpace(r.FormValue("location_name")),
        "latitude":            strings.TrimSpace(r.FormValue("latitude")),
        "longitude":           strings.TrimSpace(r.FormValue("longitude")),
        "units":               strings.TrimSpace(r.FormValue("units")),
        "time_format":         strings.TrimSpace(r.FormValue("time_format")),
        "theme":               strings.TrimSpace(r.FormValue("theme")),
        "show_seconds":        strings.TrimSpace(r.FormValue("show_seconds")),
        "show_location_name":  strings.TrimSpace(r.FormValue("show_location_name")),
    }
    if values.Get("show_seconds") == "" {
        values["show_seconds"] = "off"
    }
    if values.Get("show_location_name") == "" {
        values["show_location_name"] = "on"
    }
    for _, field := range detailFields {
        values[field.Name] = strings.TrimSpace(r.FormValue(field.Name))
    }
    if values.Get("theme") == "" {
        values["theme"] = defaultTheme
    }
    return values
}

func onOff(value bool) string {
    if value {
        return "on"
    }
    return "off"
}

func detailConfig(cfg Config) map[string]string {
    return map[string]string{
        "detail_high_low":      cfg.DetailHighLow,
        "detail_wind_speed":    cfg.DetailWindSpeed,
        "detail_wind_direction": cfg.DetailWindDir,
        "detail_humidity":      cfg.DetailHumidity,
        "detail_dew_point":     cfg.DetailDewPoint,
        "detail_pressure":      cfg.DetailPressure,
        "detail_precip_chance": cfg.DetailPrecip,
    }
}

func loadLatestConfig() Config {
    cfg, err := loadConfig()
    if err != nil {
        return getActiveConfig()
    }
    return cfg
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    cfg := loadLatestConfig()
    themeName := cfg.Theme
    theme, ok := themes[themeName]
    if !ok {
        theme = themes[defaultTheme]
    }
    timeFormat := cfg.TimeFormat
    showSeconds := cfg.ShowSeconds
    unitSymbol := allowedUnits[cfg.Units]
    if unitSymbol == "" {
        unitSymbol = allowedUnits[defaultConfig.Units]
    }
    detailCfg := detailConfig(cfg)

    current := WeatherResponse{}
    currentErr := getOpenWeatherJSON(buildEndpointURL(cfg, currentEndpoint), &current)
    if currentErr != nil {
        renderDashboardView(w, DashboardView{
            Theme:        theme,
            SettingsURL:  "/settings",
            Temp:         "--",
            Unit:         "",
            Desc:         "",
            TimeFormat:   timeFormat,
            ShowSeconds:  showSeconds,
            Error:        formatHTTPError(currentErr),
            Hint:         buildHint(currentErr),
            TodayDetails: nil,
        })
        return
    }

    forecastResp := ForecastResponse{}
    forecastErr := getOpenWeatherJSON(buildEndpointURL(cfg, forecastEndpoint), &forecastResp)
    var forecastErrorMessage string
    var forecastCards []ForecastCard
    if forecastErr != nil {
        forecastErrorMessage = formatHTTPError(forecastErr)
    } else {
        forecastCards = formatForecast(forecastResp.List, cfg.Units, unitSymbol, detailCfg)
    }

    todayDate := time.Now().UTC()
    todayEntries := filterEntriesByDate(forecastResp.List, todayDate)
    dailyHighLow, _ := collectDailyHighLow(todayEntries, &current.Main)
    precipChance := precChanceForDate(forecastResp.List, todayDate)
    todayDetails := formatTodayDetails(current, cfg.Units, unitSymbol, detailCfg, precipChance, dailyHighLow)

    city := current.Name
    if city == "" {
        city = strings.TrimSpace(cfg.LocationName)
        if city == "" {
            city = "Current location"
        }
    }
    if !cfg.ShowLocationName {
        city = ""
    }

    renderDashboardView(w, DashboardView{
        Theme:         theme,
        SettingsURL:   "/settings",
        City:          city,
        Temp:          fmt.Sprintf("%d", int(math.Round(current.Main.Temp))),
        Desc:          describeConditions(current.Weather, current.Clouds, true),
        Unit:          unitSymbol,
        Forecast:      forecastCards,
        ForecastError: forecastErrorMessage,
        TimeFormat:    timeFormat,
        ShowSeconds:   showSeconds,
        TodayDetails:  todayDetails,
    })
}

func renderDashboardView(w http.ResponseWriter, view DashboardView) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := dashboardTemplate.Execute(w, view); err != nil {
        http.Error(w, "unable to render dashboard", http.StatusInternalServerError)
        log.Printf("dashboard render error: %v", err)
    }
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
    status := ""
    if r.URL.Query().Get("saved") != "" {
        status = "Settings saved successfully."
    }
    cfg, err := loadConfig()
    loadError := ""
    if err != nil {
        cfg = getActiveConfig()
        loadError = err.Error()
        status = ""
    }

    values := defaultDetailValues(cfg)
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err == nil {
            values = formValuesFromRequest(r)
        }
        data := map[string]any{
            "openweather_api_key": values.Get("openweather_api_key"),
            "location_name":       values.Get("location_name"),
            "latitude":            values.Get("latitude"),
            "longitude":           values.Get("longitude"),
            "units":               values.Get("units"),
            "time_format":         values.Get("time_format"),
            "theme":               values.GetOrDefault("theme", defaultTheme),
            "show_seconds":        values.Get("show_seconds"),
            "show_location_name":  values.Get("show_location_name"),
        }
        for _, field := range detailFields {
            data[field.Name] = values.Get(field.Name)
        }
        normalized, normErr := validateConfig(data, false)
        if normErr != nil {
            loadError = normErr.Error()
        } else {
            if err := writeConfig(normalized); err != nil {
                loadError = err.Error()
            } else {
                setActiveConfig(normalized)
                http.Redirect(w, r, "/settings?saved=1", http.StatusSeeOther)
                return
            }
        }
    }

    if loadError != "" {
        status = ""
    }

    view := SettingsView{
        FormValues:           values,
        UnitOptions:          unitOptions,
        TimeFormatOptions:    timeFormatOptions,
        ThemeOptions:         themeOptions,
        DetailFields:         detailFields,
        DetailOptions:        detailOptions,
        Status:               status,
        Error:                loadError,
        DefaultDetailSettings: defaultDetailSettings,
        DefaultShowLocation:  onOff(defaultConfig.ShowLocationName),
        DefaultTheme:         defaultTheme,
        DashboardURL:         "/",
    }
    renderSettingsView(w, view)
}

func renderSettingsView(w http.ResponseWriter, view SettingsView) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := settingsTemplate.Execute(w, view); err != nil {
        http.Error(w, "unable to render settings", http.StatusInternalServerError)
        log.Printf("settings render error: %v", err)
    }
}

func filterEntriesByDate(entries []forecastEntry, target time.Time) []forecastEntry {
    var result []forecastEntry
    for _, entry := range entries {
        dt := time.Unix(entry.Dt, 0).UTC()
        if dt.Year() == target.Year() && dt.YearDay() == target.YearDay() {
            result = append(result, entry)
        }
    }
    return result
}

func buildHint(err error) string {
    var httpErr *httpError
    if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusUnauthorized {
        return "OpenWeather returned HTTP 401. Confirm that your API key is active, correct, and subscribed to the Personal plan endpoints, then restart the launcher."
    }
    return ""
}

func main() {
    cfg, err := ensureConfig()
    if err != nil {
        log.Fatalf("failed to ensure config: %v", err)
    }
    setActiveConfig(cfg)
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/settings", settingsHandler)
    log.Println("piWeather listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
