#ifndef __WIFI_H__
#define __WIFI_H__

#include "freertos/FreeRTOS.h"
#include "freertos/event_groups.h"

#include "esp_err.h"

#define WIFI_CONNECTED_BIT BIT0

esp_err_t
wifi_init(EventGroupHandle_t eg);

#endif /* __WIFI_H__ */

