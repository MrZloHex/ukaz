#include "esp_event.h"
#include "nvs_flash.h"

#include "wifi.h"


void
app_main()
{
    ESP_ERROR_CHECK(nvs_flash_init());
    ESP_ERROR_CHECK(esp_event_loop_create_default());
    ESP_ERROR_CHECK(wifi_init());
}
