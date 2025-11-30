package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "math"
    "net/http"
    "net/url"
    "sort"
    "strings"
    "time"
)

const (
    currentEndpoint  = "https://api.openweathermap.org/data/2.5/weather"
    forecastEndpoint = "https://api.openweathermap.org/data/2.5/forecast"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// WeatherResponse mirrors the current weather payload returned by OpenWeather.
type WeatherResponse struct {
    Name    string             `json:"name"`
    Main    mainData           `json:"main"`
    Wind    windData           `json:"wind"`
    Clouds  cloudsData         `json:"clouds"`
    Weather []weatherCondition `json:"weather"`
}

// ForecastResponse mirrors the five-day / three-hour forecast payload.
type ForecastResponse struct {
    List []forecastEntry `json:"list"`
}

type mainData struct {
    Temp     float64 `json:"temp"`
    TempMin  float64 `json:"temp_min"`
    TempMax  float64 `json:"temp_max"`
    Pressure float64 `json:"pressure"`
    Humidity float64 `json:"humidity"`
}

type windData struct {
    Speed float64 `json:"speed"`
    Deg   float64 `json:"deg"`
}

type cloudsData struct {
    All float64 `json:"all"`
}

type weatherCondition struct {
    Main        string `json:"main"`
    Description string `json:"description"`
}

type forecastEntry struct {
    Dt      int64              `json:"dt"`
    Main    mainData           `json:"main"`
    Wind    windData           `json:"wind"`
    Clouds  cloudsData         `json:"clouds"`
    Weather []weatherCondition `json:"weather"`
    Pop     *float64           `json:"pop"`
}

// ForecastCard and DetailItem provide structured data for templates.
type ForecastCard struct {
    Day     string
    Temp    int
    Desc    string
    Details []DetailItem
}

type DetailItem struct {
    Label string
    Value string
}

type httpError struct {
    StatusCode int
    Reason     string
    Message    string
}

func (e *httpError) Error() string {
    if e.Message != "" {
        return e.Message
    }
    return fmt.Sprintf("%d %s", e.StatusCode, e.Reason)
}

func formatHTTPError(err error) string {
    var httpErr *httpError
    if errors.As(err, &httpErr) {
        if httpErr.Message != "" {
            return httpErr.Message
        }
        return fmt.Sprintf("%d %s", httpErr.StatusCode, httpErr.Reason)
    }
    return err.Error()
}

func newHTTPError(resp *http.Response) error {
    reason := resp.Status
    message := reason
    var payload map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
        if val, ok := payload["message"].(string); ok && val != "" {
            message = val
        }
    }
    return &httpError{StatusCode: resp.StatusCode, Reason: reason, Message: message}
}

func getOpenWeatherJSON(url string, dst any) error {
    resp, err := httpClient.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return newHTTPError(resp)
    }
    if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
        return fmt.Errorf("unable to parse response: %w", err)
    }
    return nil
}

func buildEndpointURL(cfg Config, endpoint string) string {
    params := url.Values{}
    params.Set("lat", fmt.Sprintf("%f", cfg.Latitude))
    params.Set("lon", fmt.Sprintf("%f", cfg.Longitude))
    params.Set("appid", cfg.OpenWeatherAPIKey)
    params.Set("units", cfg.Units)
    return endpoint + "?" + params.Encode()
}

func formatForecast(entries []forecastEntry, units, unitSymbol string, detailConfig map[string]string) []ForecastCard {
    if len(entries) == 0 {
        return nil
    }

    today := time.Now().UTC()
    grouped := map[string][]forecastEntry{}
    for _, entry := range entries {
        dt := time.Unix(entry.Dt, 0).UTC()
        key := dt.Format("2006-01-02")
        grouped[key] = append(grouped[key], entry)
    }

    var days []time.Time
    for key := range grouped {
        if parsed, err := time.Parse("2006-01-02", key); err == nil {
            days = append(days, parsed)
        }
    }
    sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

    var cards []ForecastCard
    for _, day := range days {
        if !day.After(today.Truncate(24 * time.Hour)) {
            continue
        }
        dayEntries := grouped[day.Format("2006-01-02")]
        bestIdx := 0
        bestDiff := math.Inf(1)
        for idx, entry := range dayEntries {
            dt := time.Unix(entry.Dt, 0).UTC()
            diff := math.Abs(float64(dt.Hour() - 12))
            if diff < bestDiff {
                bestDiff = diff
                bestIdx = idx
            }
        }
        bestEntry := dayEntries[bestIdx]
        cards = append(cards, ForecastCard{
            Day:     day.Format("Mon"),
            Temp:    int(math.Round(bestEntry.Main.Temp)),
            Desc:    describeConditions(bestEntry.Weather, bestEntry.Clouds, true),
            Details: buildDayDetails(dayEntries, units, unitSymbol, detailConfig),
        })
        if len(cards) == 5 {
            break
        }
    }

    return cards
}

func describeConditions(conditions []weatherCondition, clouds cloudsData, short bool) string {
    base := "Conditions"
    if len(conditions) > 0 {
        info := conditions[0]
        if desc := strings.TrimSpace(info.Description); desc != "" {
            base = strings.Title(desc)
        } else if info.Main != "" {
            base = strings.Title(info.Main)
        }
    }

    coverage := clouds.All
    coverage = math.Max(0, math.Min(100, coverage))
    if coverage == 0 {
        return base
    }
    coveragePct := int(math.Round(coverage))
    descriptor := cloudDescriptor(coverage)
    if short {
        return fmt.Sprintf("%s %d%%", descriptor, coveragePct)
    }
    coverageText := fmt.Sprintf("%d%% cover", coveragePct)
    if len(conditions) > 0 && strings.EqualFold(conditions[0].Main, "Clouds") {
        return fmt.Sprintf("%s (%s)", descriptor, coverageText)
    }
    return fmt.Sprintf("%s • %s (%s)", base, strings.ToLower(descriptor), coverageText)
}

func cloudDescriptor(coverage float64) string {
    switch {
    case coverage <= 10:
        return "Clear"
    case coverage <= 35:
        return "Mostly clear"
    case coverage <= 65:
        return "Partly cloudy"
    case coverage <= 85:
        return "Mostly cloudy"
    default:
        return "Overcast"
    }
}

func buildDayDetails(entries []forecastEntry, units, unitSymbol string, detailConfig map[string]string) []DetailItem {
    var details []DetailItem

    if detailShows(detailConfig, "detail_high_low", "week") {
        var highs, lows []float64
        for _, entry := range entries {
            valueHigh := entry.Main.TempMax
            if valueHigh == 0 {
                valueHigh = entry.Main.Temp
            }
            valueLow := entry.Main.TempMin
            if valueLow == 0 {
                valueLow = entry.Main.Temp
            }
            highs = append(highs, valueHigh)
            lows = append(lows, valueLow)
        }
        if len(highs) > 0 && len(lows) > 0 {
            details = append(details, DetailItem{
                Label: "High / Low",
                Value: fmt.Sprintf("%d°%s / %d°%s", int(math.Round(maxFloat64(highs))), unitSymbol, int(math.Round(minFloat64(lows))), unitSymbol),
            })
        }
    }

    windSpeed := detailShows(detailConfig, "detail_wind_speed", "week")
    windDir := detailShows(detailConfig, "detail_wind_direction", "week")
    var speedValue string
    if windSpeed {
        var speeds []float64
        for _, entry := range entries {
            speeds = append(speeds, entry.Wind.Speed)
        }
        if len(speeds) > 0 {
            unit := "mph"
            if units != "imperial" {
                unit = "m/s"
            }
            speedValue = fmt.Sprintf("%d %s", int(math.Round(averageFloat64(speeds))), unit)
        }
    }
    var directionValue string
    if windDir {
        var sinSum, cosSum float64
        for _, entry := range entries {
            rad := entry.Wind.Deg * math.Pi / 180
            sinSum += math.Sin(rad)
            cosSum += math.Cos(rad)
        }
        if sinSum != 0 || cosSum != 0 {
            avgRad := math.Atan2(sinSum, cosSum)
            avgDeg := math.Mod(avgRad*180/math.Pi+360, 360)
            directionValue = degToCardinal(avgDeg)
        }
    }
    if (windSpeed && speedValue != "") || (windDir && directionValue != "") {
        var parts []string
        if windSpeed && speedValue != "" {
            parts = append(parts, speedValue)
        }
        if windDir && directionValue != "" {
            parts = append(parts, directionValue)
        }
        label := "Wind"
        if !windSpeed {
            label = "Wind direction"
        }
        details = append(details, DetailItem{Label: label, Value: strings.Join(parts, " • ")})
    }

    if detailShows(detailConfig, "detail_humidity", "week") {
        var humidities []float64
        for _, entry := range entries {
            humidities = append(humidities, entry.Main.Humidity)
        }
        if len(humidities) > 0 {
            details = append(details, DetailItem{Label: "Humidity", Value: fmt.Sprintf("%d%%", int(math.Round(averageFloat64(humidities))))})
        }
    }

    if detailShows(detailConfig, "detail_dew_point", "week") {
        var dewPoints []float64
        for _, entry := range entries {
            if entry.Main.Temp == 0 || entry.Main.Humidity == 0 {
                continue
            }
            dewC := calculateDewPoint(toCelsius(entry.Main.Temp, units), entry.Main.Humidity)
            if !math.IsNaN(dewC) {
                dewPoints = append(dewPoints, fromCelsius(dewC, units))
            }
        }
        if len(dewPoints) > 0 {
            details = append(details, DetailItem{Label: "Dew point", Value: fmt.Sprintf("%d°%s", int(math.Round(averageFloat64(dewPoints))), unitSymbol)})
        }
    }

    if detailShows(detailConfig, "detail_pressure", "week") {
        var pressures []float64
        for _, entry := range entries {
            pressures = append(pressures, entry.Main.Pressure)
        }
        if len(pressures) > 0 {
            details = append(details, DetailItem{Label: "Pressure", Value: formatPressure(averageFloat64(pressures), units)})
        }
    }

    if detailShows(detailConfig, "detail_precip_chance", "week") {
        var pops []float64
        for _, entry := range entries {
            if entry.Pop != nil {
                pops = append(pops, *entry.Pop)
            }
        }
        if len(pops) > 0 {
            chance := maxFloat64(pops)
            details = append([]DetailItem{{Label: "Precip", Value: fmt.Sprintf("%d%%", int(math.Round(chance*100)))}} , details...)
        }
    }

    return details
}

func detailShows(detailConfig map[string]string, key, context string) bool {
    value := detailConfig[key]
    if value == "" {
        value = defaultDetailSettings[key]
    }
    if value == "both" {
        return true
    }
    return value == context
}

func averageFloat64(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }
    total := 0.0
    for _, value := range values {
        total += value
    }
    return total / float64(len(values))
}

func maxFloat64(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }
    max := values[0]
    for _, value := range values {
        if value > max {
            max = value
        }
    }
    return max
}

func minFloat64(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }
    min := values[0]
    for _, value := range values {
        if value < min {
            min = value
        }
    }
    return min
}

func degToCardinal(degrees float64) string {
    dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
    idx := int(math.Mod(degrees+11.25, 360) / 22.5)
    return dirs[idx%len(dirs)]
}

func calculateDewPoint(tempC, humidity float64) float64 {
    if tempC == 0 && humidity <= 0 {
        return math.NaN()
    }
    a, b := 17.625, 243.04
    alpha := ((a * tempC) / (b + tempC)) + math.Log(humidity/100.0)
    return (b * alpha) / (a - alpha)
}

func toCelsius(temp float64, units string) float64 {
    if units == "imperial" {
        return (temp - 32) * 5 / 9
    }
    return temp
}

func fromCelsius(temp float64, units string) float64 {
    if units == "imperial" {
        return temp*9/5 + 32
    }
    return temp
}

func formatPressure(pressure float64, units string) string {
    if units == "imperial" {
        return fmt.Sprintf("%.2f inHg", pressure*0.02953)
    }
    return fmt.Sprintf("%d hPa", int(math.Round(pressure)))
}

func precChanceForDate(entries []forecastEntry, target time.Time) *float64 {
    var pops []float64
    for _, entry := range entries {
        dt := time.Unix(entry.Dt, 0).UTC()
        if dt.Year() == target.Year() && dt.YearDay() == target.YearDay() {
            if entry.Pop != nil {
                pops = append(pops, *entry.Pop)
            }
        }
    }
    if len(pops) == 0 {
        return nil
    }
    chance := maxFloat64(pops)
    return &chance
}

func collectDailyHighLow(entries []forecastEntry, current *mainData) (*[2]float64, bool) {
    var highs, lows []float64
    for _, entry := range entries {
        valueHigh := entry.Main.TempMax
        valueLow := entry.Main.TempMin
        if valueHigh == 0 {
            valueHigh = entry.Main.Temp
        }
        if valueLow == 0 {
            valueLow = entry.Main.Temp
        }
        highs = append(highs, valueHigh)
        lows = append(lows, valueLow)
    }
    if current != nil {
        if current.TempMax != 0 {
            highs = append(highs, current.TempMax)
        } else {
            highs = append(highs, current.Temp)
        }
        if current.TempMin != 0 {
            lows = append(lows, current.TempMin)
        } else {
            lows = append(lows, current.Temp)
        }
    }
    if len(highs) == 0 || len(lows) == 0 {
        return nil, false
    }
    result := &[2]float64{maxFloat64(highs), minFloat64(lows)}
    return result, true
}

func formatTodayDetails(current WeatherResponse, units, unitSymbol string, detailConfig map[string]string, precipChance *float64, dailyHighLow *[2]float64) []DetailItem {
    var details []DetailItem
    main := current.Main
    wind := current.Wind

    if detailShows(detailConfig, "detail_high_low", "today") {
        if dailyHighLow != nil {
            details = append(details, DetailItem{
                Label: "High / Low",
                Value: fmt.Sprintf("%d°%s / %d°%s", int(math.Round(dailyHighLow[0])), unitSymbol, int(math.Round(dailyHighLow[1])), unitSymbol),
            })
        } else if main.TempMax != 0 || main.TempMin != 0 {
            high := main.TempMax
            if high == 0 {
                high = main.Temp
            }
            low := main.TempMin
            if low == 0 {
                low = main.Temp
            }
            details = append(details, DetailItem{
                Label: "High / Low",
                Value: fmt.Sprintf("%d°%s / %d°%s", int(math.Round(high)), unitSymbol, int(math.Round(low)), unitSymbol),
            })
        }
    }

    windSpeed := detailShows(detailConfig, "detail_wind_speed", "today")
    windDir := detailShows(detailConfig, "detail_wind_direction", "today")
    windParts := []string{}
    if windSpeed && wind.Speed != 0 {
        unit := "mph"
        if units != "imperial" {
            unit = "m/s"
        }
        windParts = append(windParts, fmt.Sprintf("%d %s", int(math.Round(wind.Speed)), unit))
    }
    if windDir && (wind.Deg != 0 || wind.Speed != 0) {
        windParts = append(windParts, degToCardinal(wind.Deg))
    }
    if len(windParts) > 0 {
        label := "Wind"
        if !windSpeed {
            label = "Wind direction"
        }
        details = append(details, DetailItem{Label: label, Value: strings.Join(windParts, " • ")})
    }

    if detailShows(detailConfig, "detail_humidity", "today") && main.Humidity != 0 {
        details = append(details, DetailItem{Label: "Humidity", Value: fmt.Sprintf("%d%%", int(math.Round(main.Humidity)))})
    }

    if detailShows(detailConfig, "detail_dew_point", "today") && main.Temp != 0 && main.Humidity > 0 {
        dewC := calculateDewPoint(toCelsius(main.Temp, units), main.Humidity)
        if !math.IsNaN(dewC) {
            details = append(details, DetailItem{Label: "Dew point", Value: fmt.Sprintf("%d°%s", int(math.Round(fromCelsius(dewC, units))), unitSymbol)})
        }
    }

    if detailShows(detailConfig, "detail_pressure", "today") && main.Pressure != 0 {
        details = append(details, DetailItem{Label: "Pressure", Value: formatPressure(main.Pressure, units)})
    }

    if detailShows(detailConfig, "detail_precip_chance", "today") && precipChance != nil {
        details = append(details, DetailItem{Label: "Precip chance", Value: fmt.Sprintf("%d%%", int(math.Round(*precipChance*100)))})
    }

    return details
}
