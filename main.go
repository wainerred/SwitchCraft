package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	BluePort    string `json:"blue_port"`
	GreenPort   string `json:"green_port"`
	ProxyPort   string `json:"proxy_port"`
	Current     string `json:"current"`
	ServiceName string `json:"service_name"`
}

type DeploymentStatus struct {
	BlueHealthy  bool      `json:"blue_healthy"`
	GreenHealthy bool      `json:"green_healthy"`
	LastChecked  time.Time `json:"last_checked"`
	BlueVersion  string    `json:"blue_version"`
	GreenVersion string    `json:"green_version"`
}

var (
	config Config
	status DeploymentStatus
)

func main() {
	loadConfig()
	
	go healthChecker()

	// Web UI routes
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/api/switch", switchHandler)
	http.HandleFunc("/api/status", apiStatusHandler)
	http.HandleFunc("/api/config", configHandler)
	http.HandleFunc("/api/deploy", deployHandler)
	
	// Proxy route - все остальные запросы проксируются к активной среде
	http.HandleFunc("/", proxyHandler)

	log.Printf("🚀 Blue-Green Manager started on :%s", config.ProxyPort)
	log.Printf("📊 Service: %s", config.ServiceName)
	log.Fatal(http.ListenAndServe(":"+config.ProxyPort, nil))
}

func loadConfig() {
	config = Config{
		BluePort:    "5176",
		GreenPort:   "5177", 
		ProxyPort:   "8080",
		Current:     "blue",
		ServiceName: "Frontend Application",
	}

	if port := os.Getenv("BLUE_PORT"); port != "" {
		config.BluePort = port
	}
	if port := os.Getenv("GREEN_PORT"); port != "" {
		config.GreenPort = port
	}
	if port := os.Getenv("PROXY_PORT"); port != "" {
		config.ProxyPort = port
	}
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		config.ServiceName = name
	}
}

func healthChecker() {
	for {
		status.BlueHealthy = checkHealth(config.BluePort)
		status.GreenHealthy = checkHealth(config.GreenPort)
		status.LastChecked = time.Now()
		
		// Получаем версии приложений
		status.BlueVersion = getVersion(config.BluePort)
		status.GreenVersion = getVersion(config.GreenPort)
		
		time.Sleep(10 * time.Second)
	}
}

func checkHealth(port string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://app-" + port + ":" + port + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getVersion(port string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://app-" + port + ":" + port + "/version")
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()
	
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result["version"]
}

// Web Interface
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		proxyHandler(w, r)
		return
	}

	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Blue-Green Deployment</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold text-gray-800 mb-2">🚀 {{.ServiceName}}</h1>
        <p class="text-gray-600 mb-8">Blue-Green Deployment Manager</p>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
            <!-- Blue Environment -->
            <div class="bg-white rounded-lg shadow p-6 {{if eq .Current "blue"}}border-2 border-blue-500{{end}}">
                <div class="flex items-center justify-between mb-4">
                    <h2 class="text-xl font-semibold text-blue-600">Blue Environment</h2>
                    <span class="px-3 py-1 rounded-full text-sm {{if .BlueHealthy}}bg-green-100 text-green-800{{else}}bg-red-100 text-red-800{{end}}">
                        {{if .BlueHealthy}}Healthy{{else}}Unhealthy{{end}}
                    </span>
                </div>
                <div class="space-y-2">
                    <p><strong>Port:</strong> {{.BluePort}}</p>
                    <p><strong>Version:</strong> {{.BlueVersion}}</p>
                    <p><strong>Status:</strong> {{if eq .Current "blue"}}<span class="text-green-600 font-semibold">ACTIVE</span>{{else}}Standby{{end}}</p>
                </div>
            </div>

            <!-- Green Environment -->
            <div class="bg-white rounded-lg shadow p-6 {{if eq .Current "green"}}border-2 border-green-500{{end}}">
                <div class="flex items-center justify-between mb-4">
                    <h2 class="text-xl font-semibold text-green-600">Green Environment</h2>
                    <span class="px-3 py-1 rounded-full text-sm {{if .GreenHealthy}}bg-green-100 text-green-800{{else}}bg-red-100 text-red-800{{end}}">
                        {{if .GreenHealthy}}Healthy{{else}}Unhealthy{{end}}
                    </span>
                </div>
                <div class="space-y-2">
                    <p><strong>Port:</strong> {{.GreenPort}}</p>
                    <p><strong>Version:</strong> {{.GreenVersion}}</p>
                    <p><strong>Status:</strong> {{if eq .Current "green"}}<span class="text-green-600 font-semibold">ACTIVE</span>{{else}}Standby{{end}}</p>
                </div>
            </div>
        </div>

        <!-- Controls -->
        <div class="bg-white rounded-lg shadow p-6 mb-6">
            <h3 class="text-lg font-semibold mb-4">Deployment Controls</h3>
            <div class="flex space-x-4">
                <button onclick="switchEnvironment()" 
                    class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-lg transition disabled:opacity-50"
                    {{if not (and .BlueHealthy .GreenHealthy)}}disabled{{end}}>
                    Switch to {{if eq .Current "blue"}}Green{{else}}Blue{{end}}
                </button>
                
                <button onclick="refreshStatus()" class="bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg transition">
                    Refresh Status
                </button>
            </div>
        </div>

        <!-- Configuration -->
        <div class="bg-white rounded-lg shadow p-6">
            <h3 class="text-lg font-semibold mb-4">Configuration</h3>
            <form onsubmit="updateConfig(event)" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Blue Port</label>
                    <input type="text" id="bluePort" value="{{.BluePort}}" class="w-full px-3 py-2 border border-gray-300 rounded-md">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Green Port</label>
                    <input type="text" id="greenPort" value="{{.GreenPort}}" class="w-full px-3 py-2 border border-gray-300 rounded-md">
                </div>
                <div class="md:col-span-2">
                    <button type="submit" class="bg-purple-500 hover:bg-purple-600 text-white px-6 py-2 rounded-lg transition">
                        Update Configuration
                    </button>
                </div>
            </form>
        </div>
    </div>

    <script>
        async function switchEnvironment() {
            const response = await fetch('/api/switch', { method: 'POST' });
            const result = await response.json();
            if (result.status === 'success') {
                alert('Environment switched to ' + result.current);
                location.reload();
            } else {
                alert('Switch failed: ' + result.error);
            }
        }

        async function updateConfig(event) {
            event.preventDefault();
            const config = {
                blue_port: document.getElementById('bluePort').value,
                green_port: document.getElementById('greenPort').value
            };
            
            const response = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            
            const result = await response.json();
            if (result.status === 'success') {
                alert('Configuration updated');
                location.reload();
            }
        }

        function refreshStatus() {
            location.reload();
        }

        // Auto-refresh every 30 seconds
        setInterval(refreshStatus, 30000);
    </script>
</body>
</html>`

	tpl, _ := template.New("dashboard").Parse(tmpl)
	data := struct {
		Config
		DeploymentStatus
	}{config, status}
	
	w.Header().Set("Content-Type", "text/html")
	tpl.Execute(w, data)
}

// API Handlers
func switchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем что целевая среда здорова
	target := "green"
	if config.Current == "blue" {
		target = "blue"
	}

	var healthy bool
	if target == "blue" {
		healthy = status.BlueHealthy
	} else {
		healthy = status.GreenHealthy
	}

	if !healthy {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error", 
			"error": "Target environment is not healthy",
		})
		return
	}

	// Переключаем
	old := config.Current
	config.Current = target
	saveConfig()

	log.Printf("🔄 Switched from %s to %s", old, config.Current)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"old":     old,
		"current": config.Current,
	})
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"config": config,
		"status": status,
	})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
		return
	}

	if r.Method == "POST" {
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		config.BluePort = newConfig.BluePort
		config.GreenPort = newConfig.GreenPort
		saveConfig()

		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"message": "Configuration updated",
		})
	}
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	// Здесь можно добавить логику для автоматического деплоя
	// Например, запуск Docker Compose или вызов CI/CD API
	
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Deployment triggered",
	})
}

// Proxy Handler - проксирует запросы к активной среде
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Исключаем API и статические пути из проксирования
	if r.URL.Path == "/" || r.URL.Path == "/api/switch" || r.URL.Path == "/api/status" || 
	   r.URL.Path == "/api/config" || r.URL.Path == "/api/deploy" {
		return
	}

	var targetPort string
	if config.Current == "blue" {
		targetPort = config.BluePort
	} else {
		targetPort = config.GreenPort
	}

	targetURL := fmt.Sprintf("http://app-%s:%s", targetPort, targetPort)
	url, _ := url.Parse(targetURL)
	
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
}

func saveConfig() {
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("config.json", data, 0644)
}
