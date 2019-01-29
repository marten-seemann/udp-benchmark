#include <sys/socket.h>
#include <iostream>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <map>
#include <chrono>
#include <fstream>
#include "../config.h"

int main(int argc, const char * argv[]) {
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Opening datagram socket");
        exit(1);
    }
    
    struct sockaddr_in servaddr;
    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET; //we're using inet
    servaddr.sin_addr.s_addr = INADDR_ANY;
    servaddr.sin_port = htons(PROXY_PORT); //set the port
    
    std::chrono::high_resolution_clock clock;
    std::map<uint64_t /* packet number */, uint64_t /* timestamp in ns */> sendTimes;
    
    for(uint64_t pn = 0; pn < NUM_PACKETS; pn++) {
        unsigned char value[sizeof(pn)];
        std::memcpy(value, &pn, sizeof(pn)); // binary.LittleEndian.PutUint64()
        if(sendto(sock, (const char *) value, MSGLEN, 0, (const struct sockaddr *) &servaddr, sizeof(servaddr)) < 0) {
            perror("Error sending packet");
            exit(1);
        }
        sendTimes[pn] = std::chrono::duration_cast<std::chrono::nanoseconds>(clock.now().time_since_epoch()).count();
        usleep(100);
    }

    std::ofstream outdata; 
    outdata.open("sendtimes.txt"); // opens the file
    for(auto it = sendTimes.begin(); it != sendTimes.end(); ++it) {
        outdata << it->first << " " << it->second << std::endl;
    }
    outdata.close();
    close(sock);
}
