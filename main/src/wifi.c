#include "wifi.h"

#include "esp_log.h"
#include "tcpip_adapter.h"
#include "esp_wifi.h"


#define ERR_CHECK_RET(__X__) ({                         \
        esp_err_t __err_ret = (__X__);                  \
        ESP_ERROR_CHECK_WITHOUT_ABORT(__err_ret);       \
        if (__err_ret) { return __err_ret; } })


static const char *TAG = "wifi";
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
                ESP_LOGI(TAG, "got ip: %s", ip4addr_ntoa(&event->ip_info.ip));
                s_retry_num = 0;
                // xEventGroupSetBits(s_wifi_event_group, WIFI_CONNECTED_BIT);
            } break;
        }
    }
}

esp_err_t
wifi_init(void)
{
    ESP_LOGI(TAG, "initializing");
    tcpip_adapter_init();

    ERR_CHECK_RET(tcpip_adapter_set_hostname(TCPIP_ADAPTER_IF_STA, CONFIG_NODE_NAME));

    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    ERR_CHECK_RET(esp_wifi_init(&cfg));
    ERR_CHECK_RET(esp_wifi_set_mode(WIFI_MODE_STA));


    ERR_CHECK_RET(esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID, &wifi_handler, NULL));
    ERR_CHECK_RET(esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP, &wifi_handler, NULL));

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

    return ESP_OK;
}
