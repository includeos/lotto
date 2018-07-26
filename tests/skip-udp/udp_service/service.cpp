// This file is a part of the IncludeOS unikernel - www.includeos.org
//
// Copyright 2018 IncludeOS AS, Oslo, Norway
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <service>
#include <net/inet>
#include <timers>

using namespace net;

void Service::start() {
    auto& inet = net::Super_stack::get(0);
    const UDP::port_t port = 4242;
    auto& sock = inet.udp().bind(port);

    // When getting UDP data
    sock.on_read([&sock] (UDP::addr_t addr, UDP::port_t port,
        const char* data, size_t len)
    {
        std::string strdata(data, len);
        INFO("UDP service", "Getting UDP data from %s:%d %s",
            addr.to_string().c_str(), port, strdata.c_str());

        // Send the same data right back
        using namespace std::chrono;
        Timers::oneshot(100ms, Timers::handler_t::make_packed([&sock, addr, port, data, len](Timers::id_t) {
            INFO("UDP service", "Sending UDP reply to %s:%d", addr.to_string().c_str(), port);
            sock.sendto(addr, port, data, len);
        }));
    });

    INFO("UDP service", "Listening on port %d", port);
}
