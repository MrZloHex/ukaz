#include "net/wifi.h"

#include "esp_log.h"
#include "tcpip_adapter.h"
#include "esp_wifi.h"

#include "utils.h"



static const char *TAG = "WIFI";
static uint8_t s_retry_num = 0;

static void
wifi_handler
(
    void *arg,
    esp_event_base_t event_base,
    int32_t event_id,
    void   *event_data
)
{
    if (event_base == WIFI_EVENT)
    {
        switch (event_id)
        {
            case WIFI_EVENT_STA_START:
            {
                ESP_ERROR_CHECK_WITHOUT_ABORT(tcpip_adapter_set_hostname(TCPIP_ADAPTER_IF_STA, CONFIG_NODE_NAME));
                ESP_LOGI(TAG, "connecting");
                ESP_ERROR_CHECK_WITHOUT_ABORT(esp_wifi_connect());
            } break;

            case WIFI_EVENT_STA_DISCONNECTED:
            {
                if (s_retry_num < CONFIG_ESP_MAXIMUM_RETRY)
                {
                    ESP_LOGI(TAG, "retry to connect to the AP");
                    ESP_ERROR_CHECK_WITHOUT_ABORT(esp_wifi_connect());
                    s_retry_num += 1;
                } 
                else 
                { /* xEventGroupSetBits(s_wifi_event_group, WIFI_FAIL_BIT); */ }
                ESP_LOGI(TAG,"connect to the AP fail");
            } break;
        }
    }
    else if (event_base == IP_EVENT)
    {
        switch (event_id)
        {
            case IP_EVENT_STA_GOT_IP:
            {
                ip_event_got_ip_t *event = (ip_event_got_ip_t *)event_data;
                ESP_LOGI
                (
                    TAG,
                    "got ip: " IPSTR ", mask: " IPSTR ", gw: " IPSTR,
                    IP2STR(&event->ip_info.ip),
                    IP2STR(&event->ip_info.netmask),
                    IP2STR(&event->ip_info.gw)
                );
                s_retry_num = 0;

                EventGroupHandle_t eg = (EventGroupHandle_t)arg;
                xEventGroupSetBits(eg, WIFI_CONNECTED_BIT);
            } break;
        }
    }
}

esp_err_t
wifi_init(EventGroupHandle_t eg)
{
    ESP_LOGI(TAG, "init");
    tcpip_adapter_init();

    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    ERR_CHECK_RET(esp_wifi_init(&cfg));
    ERR_CHECK_RET(esp_wifi_set_mode(WIFI_MODE_STA));


    ERR_CHECK_RET(esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID, &wifi_handler, eg));
    ERR_CHECK_RET(esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP, &wifi_handler, eg));

    wifi_config_t wifi_config =
    {
        .sta =
        {
            .ssid               = CONFIG_ESP_WIFI_SSID,
            .password           = CONFIG_ESP_WIFI_PASSWORD,
            .threshold.authmode = WIFI_AUTH_WPA2_PSK
        },
    };
    ERR_CHECK_RET(esp_wifi_set_config(ESP_IF_WIFI_STA, &wifi_config));
    ERR_CHECK_RET(esp_wifi_start());

    ESP_LOGI(TAG, "OK");

    return ESP_OK;
}
