cmake_minimum_required(VERSION 2.8.9)

# IncludeOS install location
if (NOT DEFINED ENV{INCLUDEOS_PREFIX})
  set(ENV{INCLUDEOS_PREFIX} /usr/local)
endif()

# Use toolchain (if needed)
include($ENV{INCLUDEOS_PREFIX}/includeos/pre.service.cmake)

include(dependencies.cmake)

# Name of your project
project (acorn_service)

# Human-readable name of your service
set(SERVICE_NAME "IncludeOS Acorn Web Service")

# Name of your service binary
set(BINARY       "acorn_service")

# Source files to be linked with OS library parts to form bootable image
set(SOURCES
  service.cpp fs/acorn_fs.cpp
  )

#
# Service CMake options
# (uncomment to enable)
#

# MISC:

# To add your own include paths:
set(LOCAL_INCLUDES
  "app/"
  ${BUCKET_DIR}
  )

# Adding memdisk (expects my.disk to exist in current dir):
# set(MEMDISK ${CMAKE_SOURCE_DIR}/my.disk)

# DRIVERS / PLUGINS:

set(DRIVERS
  virtionet
  vmxnet3
  uplink_log
  boot_logger
  )

set(PLUGINS
  uplink
  autoconf
  system_log
  )

set(STDOUT
  vga_output
  )

# THIRD PARTY LIBRARIES:

set(LIBRARIES
  libliveupdate.a libmana.a # path to full library
  )

# include service build script
include($ENV{INCLUDEOS_PREFIX}/includeos/post.service.cmake)
install_certificates(disk/certs)
diskbuilder(disk)

# Make sure bucket is downloaded before service is built
add_dependencies(service bucket)
