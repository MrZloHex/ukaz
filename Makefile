# # ==============================================================================
#
#	 ███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
#	 ████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
#	 ██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
#	 ██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
#	 ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
#	 ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝
#
#                           ░▒▓█ _UKAZ_ █▓▒░
#
#   Makefile
#   Author     : MrZloHex
#   Date       : 2026-04-15
#   Version    : 1.0
#
#   Description:
#       This Makefile compiles and links the UKAZ project sources.
#
#   Warning    : This Makefile is so cool it might make your terminal shine!
# ==============================================================================


PROJECT_NAME := ukaz
EXCLUDE_COMPONENTS := freemodbus fatfs

include $(IDF_PATH)/make/project.mk

CERTS_DIR    := $(strip $(subst ",,$(CONFIG_CERTS_DIR)))
SPIFFS_DIR   := $(PROJECT_PATH)/$(CERTS_DIR)
MKSPIFFS     := $(strip $(subst ",,$(CONFIG_MKSPIFFS_BIN)))
SPIFFS_BIN   := $(strip $(subst ",,$(CONFIG_SPIFFS_PARTITION_BIN)))
SPIFFS_BLOCK := $(CONFIG_SPIFFS_BLOCK_SIZE)
SPIFFS_PAGE  := $(CONFIG_SPIFFS_PAGE_SIZE)
SPIFFS_SIZE  := $(CONFIG_SPIFFS_STORAGE_SIZE)
SPIFFS_OFFSET:= $(CONFIG_SPIFFS_STORAGE_OFFSET)
PORT         := $(strip $(subst ",,$(CONFIG_ESPTOOLPY_PORT)))
BAUD         := $(CONFIG_ESPTOOLPY_BAUD)
BEFORE       := $(strip $(subst ",,$(CONFIG_ESPTOOLPY_BEFORE)))
AFTER        := $(strip $(subst ",,$(CONFIG_ESPTOOLPY_AFTER)))

MKSPIFFS_REQUIRED_FLAGS := -DSPIFFS_OBJ_NAME_LEN=32 -DSPIFFS_OBJ_META_LEN=4 -DSPIFFS_USE_MAGIC=1 -DSPIFFS_USE_MAGIC_LENGTH=1 -DSPIFFS_ALIGNED_OBJECT_INDEX_TABLES=4

.PHONY: spiffs flash-spiffs check-mkspiffs-version

check-mkspiffs-version:
	@MKSPIFFS_OUT=$$($(MKSPIFFS) --version 2>&1); \
	FLAGS=$$(echo "$$MKSPIFFS_OUT" | grep '^Extra build flags:' | sed 's/^Extra build flags: //'); \
	if [ "$$FLAGS" != "$(MKSPIFFS_REQUIRED_FLAGS)" ]; then \
		echo "ERROR: mkspiffs build flags mismatch."; \
		echo "  Required : $(MKSPIFFS_REQUIRED_FLAGS)"; \
		echo "  Found    : $$FLAGS"; \
		exit 1; \
	fi; \
	echo "mkspiffs OK (flags match)"

spiffs: check-mkspiffs-version
	$(MKSPIFFS) -c $(SPIFFS_DIR) -b $(SPIFFS_BLOCK) -p $(SPIFFS_PAGE) -s $(SPIFFS_SIZE) $(SPIFFS_BIN)

flash-spiffs: $(SPIFFS_BIN)
	esptool.py --chip esp8266 --port $(PORT) --baud $(BAUD) \
		--before $(BEFORE) --after $(AFTER) \
		write_flash $(SPIFFS_OFFSET) $(SPIFFS_BIN)

