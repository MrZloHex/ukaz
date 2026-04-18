#ifndef __UTILS_H__
#define __UTILS_H__

#define ERR_CHECK_RET(__X__) ({                         \
        esp_err_t __err_ret = (__X__);                  \
        ESP_ERROR_CHECK_WITHOUT_ABORT(__err_ret);       \
        if (__err_ret) { return __err_ret; } })

#endif /* __UTILS_H__ */
