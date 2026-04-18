███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝


  ░▒▓█ _ukaz_ █▓▒░



void
esp_transport_set_errors(esp_transport_handle_t t, const esp_tls_error_handle_t error_handle)
{
    if (t != NULL && t->error_handle != NULL && error_handle != NULL)
    {
        memcpy(t->error_handle, error_handle, sizeof(esp_tls_last_error_t));
    }
}

Build configuration name: esp8266rtos
SPIFFS ver. 0.3.7-5-gf5e26c4
Extra build flags: -DSPIFFS_OBJ_NAME_LEN=32 -DSPIFFS_OBJ_META_LEN=4 -DSPIFFS_USE_MAGIC=1 -DSPIFFS_USE_MAGIC_LENGTH=1 -DSPIFFS_ALIGNED_OBJECT_INDEX_TABLES=4
SPIFFS configuration:
  SPIFFS_OBJ_NAME_LEN: 32
  SPIFFS_OBJ_META_LEN: 4
  SPIFFS_USE_MAGIC: 1
  SPIFFS_USE_MAGIC_LENGTH: 1
  SPIFFS_ALIGNED_OBJECT_INDEX_TABLES: 4
