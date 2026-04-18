#include "network.h"

#include "esp_log.h"
#include "esp_transport_ssl.h"
#include "esp_transport_ws.h"

#include "net/wifi.h"
#include "net/certs.h"
#include "utils.h"


static const char *TAG = "NETWORK";

esp_err_t
network_init(Network *net)
{
    ESP_LOGI(TAG, "init");

    net->wifi_eg = xEventGroupCreate();
    ERR_CHECK_RET(wifi_init(net->wifi_eg));

    net->list = esp_transport_list_init();
    if (!net->list)
    { ESP_LOGE(TAG, "failed to init NET LIST"); return ESP_FAIL; }

    net->ssl = esp_transport_ssl_init();
    if (!net->ssl)
    { ESP_LOGE(TAG, "failed to init SSL"); return ESP_FAIL; }

    ERR_CHECK_RET(certs_init(&net->certs));
    ERR_CHECK_RET(certs_load(&net->certs, net->ssl));

#if CONFIG_CONC_URI_IP == 1
    esp_transport_ssl_skip_common_name_check(net->ssl);
#endif

    net->wss = esp_transport_ws_init(net->ssl);
    if (!net->wss)
    { ESP_LOGE(TAG, "failed to init WS"); return ESP_FAIL; }

    esp_transport_ws_set_path(net->wss, "/");

    ERR_CHECK_RET(esp_transport_list_add(net->list, net->ssl, "ssl"));
    ERR_CHECK_RET(esp_transport_list_add(net->list, net->wss, "wss"));

    ESP_LOGI(TAG, "SSL and WS initialized OK");

    xEventGroupWaitBits
    (
        net->wifi_eg,
        WIFI_CONNECTED_BIT,
        pdFALSE,
        pdTRUE,
        portMAX_DELAY
    );

    if
    (
        esp_transport_connect
        (
            net->wss, CONFIG_CONCENTRATOR_URI,
            CONFIG_CONCENTRATOR_PORT, 10000
        ) < 0
    )
    {
        ESP_LOGE(TAG, "connect failed, errno=%d", errno);
        return ESP_FAIL;
    }

    ESP_LOGI(TAG, "init OK");

    return ESP_OK;
}
