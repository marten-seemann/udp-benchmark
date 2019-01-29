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
    
    struct sockaddr_in servaddr, cliaddr;
    
    memset(&servaddr, 0, sizeof(servaddr));
    memset(&cliaddr, 0, sizeof(cliaddr));
    
    /* Bind our local address so that the client can send to us */
    bzero((char *) &servaddr, sizeof(servaddr));
    servaddr.sin_family = AF_INET; // ipv4
    servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
    servaddr.sin_port = htons(SERVER_PORT);
    
    if (bind(sock, (struct sockaddr *) &servaddr, sizeof(servaddr)) != 0) {
        perror("binding datagram socket");
        exit(1);
    }
    
    printf("Socket has port number %d\n", ntohs(servaddr.sin_port));

    std::chrono::high_resolution_clock clock;
    std::map<uint64_t /* packet number */, uint64_t /* timestamp in ns */> recvTimes;

    for(int i = 0; i < NUM_PACKETS; i++) {
        int len;
        char message[MSGLEN];
        int n = recvfrom(sock, (char *)message, MSGLEN, MSG_WAITALL, (struct sockaddr*) &cliaddr, (unsigned int*) &len);
        if (n <= 0) {
            exit(1);
        }
        uint64_t pn;
        memcpy(&pn, message, sizeof(message));

        recvTimes[pn] = std::chrono::duration_cast<std::chrono::nanoseconds>(clock.now().time_since_epoch()).count();
    }

    std::ofstream outdata; 
    outdata.open("recvtimes.txt"); // opens the file
    for(auto it = recvTimes.begin(); it != recvTimes.end(); ++it) {
        outdata << it->first << " " << it->second << std::endl;
    }
    outdata.close();
}
