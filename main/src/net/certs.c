#include "net/certs.h"

#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include "esp_spiffs.h"
#include "esp_log.h"

#include "utils.h"

static const char *TAG = "CERTS";


static esp_err_t
read_text_file(char **data, const char *path)
{
    struct stat st;
    if (stat(path, &st) != 0)
    { return ESP_ERR_NOT_FOUND; }

    FILE *f = fopen(path, "rb");
    if (!f)
    { return ESP_ERR_NOT_SUPPORTED; }

    *data = (char *)malloc((size_t)st.st_size + 1);
    if (!(*data))
    {
        fclose(f);
        return ESP_ERR_NO_MEM;
    }

    size_t n = fread(*data, 1, (size_t)st.st_size, f);
    if (n != (size_t)st.st_size)
    {
        free(*data);
        fclose(f);
        return ESP_FAIL;
    }

    (*data)[st.st_size] = 0;
    fclose(f);

    return ESP_OK;
}


esp_err_t
certs_init(mTLS_Certs *certs)
{
    esp_vfs_spiffs_conf_t spiffs =
    {
        .base_path              = "/spiffs",
        .partition_label        = NULL,
        .max_files              = 5,
        .format_if_mount_failed = true
    };

    ERR_CHECK_RET(esp_vfs_spiffs_register(&spiffs));
    ESP_LOGI(TAG,"spiffs registered");

    size_t total = 0, used = 0;
    ESP_ERROR_CHECK(esp_spiffs_info(NULL, &total, &used));
    ESP_LOGI(TAG, "Partition size: total: %d, used: %d", total, used);

    ERR_CHECK_RET(read_text_file(&certs->ca_bundle, "/spiffs/"CONFIG_CA_BUNDLE));
    ERR_CHECK_RET(read_text_file(&certs->cli_cert,  "/spiffs/"CONFIG_CLIENT_CERT));
    ERR_CHECK_RET(read_text_file(&certs->cli_key,   "/spiffs/"CONFIG_CLIENT_KEY));

    ESP_LOGI(TAG, "read OK");

    ERR_CHECK_RET(esp_vfs_spiffs_unregister(NULL));
    return ESP_OK;
}

esp_err_t
certs_load(mTLS_Certs *certs, esp_transport_handle_t ssl)
{
    size_t la = strlen(certs->ca_bundle);
    size_t lc = strlen(certs->cli_cert);
    size_t lk = strlen(certs->cli_key);

    esp_transport_ssl_set_cert_data       (ssl, certs->ca_bundle, la);
    esp_transport_ssl_set_client_cert_data(ssl, certs->cli_cert,  lc);
    esp_transport_ssl_set_client_key_data (ssl, certs->cli_key,   lk);

    ESP_LOGI(TAG, "load OK");

    return ESP_OK;
}
