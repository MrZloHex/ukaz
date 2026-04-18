#include "esp_event.h"
#include "nvs_flash.h"


#include <stdint.h>

static int
connect_wss_mtls(void)
{

#if 0
    {
        const char *msg = "{\"hello\":\"world\"}";
        int         wr;

        /* For TEXT frames, use send_raw with TEXT opcode */
        wr = esp_transport_ws_send_raw(
            ws,
            WS_TRANSPORT_OPCODES_TEXT,
            msg,
            strlen(msg),
            5000
        );

        if (wr < 0)
        {
            ESP_LOGE(TAG, "send failed, errno=%d", errno);
        }
        else
        {
            ESP_LOGI(TAG, "sent %d bytes", wr);
        }
    }

    {
        char                    rx[256];
        int                     rd;
        ws_transport_opcodes_t  opcode;

        rd = esp_transport_read(ws, rx, sizeof(rx) - 1, 10000);
        if (rd > 0)
        {
            rx[rd] = '\0';
            opcode = esp_transport_ws_get_read_opcode(ws);

            ESP_LOGI(TAG, "received %d bytes, opcode=%d", rd, (int)opcode);
            ESP_LOGI(TAG, "payload: %s", rx);
        }
        else
        {
            ESP_LOGW(TAG, "read returned %d, errno=%d", rd, errno);
        }
    }

    esp_transport_close(ws);
    esp_transport_destroy(ws);
#endif
    return 0;
}



#include "network.h"

void
app_main()
{
    ESP_ERROR_CHECK(nvs_flash_init());
    ESP_ERROR_CHECK(esp_event_loop_create_default());

    Network net = { 0 };
    ESP_ERROR_CHECK(network_init(&net));

    // connect_wss_mtls();
}
