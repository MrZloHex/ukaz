#ifndef __NETWORK_H__
#define __NETWORK_H__

#include "freertos/FreeRTOS.h"
#include "freertos/event_groups.h"

#include "esp_err.h"
#include "esp_transport.h"

#include "net/certs.h"

typedef struct
{
    EventGroupHandle_t wifi_eg;

    mTLS_Certs certs;

    esp_transport_list_handle_t list;
    esp_transport_handle_t ssl;
    esp_transport_handle_t wss;
} Network;

esp_err_t
network_init(Network *net);

#endif /* __NETWORK_H__ */
