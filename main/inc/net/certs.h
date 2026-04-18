#ifndef __CERTS_H__
#define __CERTS_H__

#include "esp_err.h"
#include "esp_transport_ssl.h"

typedef struct
{
    char *ca_bundle;
    char *cli_key;
    char *cli_cert;
} mTLS_Certs;

esp_err_t
certs_init(mTLS_Certs *certs);

esp_err_t
certs_load(mTLS_Certs *certs, esp_transport_handle_t ssl);

#endif /* __CERTS_H__ */
