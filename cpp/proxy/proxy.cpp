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
#include <dispatch/dispatch.h>
#include "../config.h"



int main(int argc, const char * argv[]) {
    dispatch_queue_t queue = dispatch_queue_create("stuff", NULL);
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Opening datagram socket");
        exit(1);
    }
    
    struct sockaddr_in proxyaddr, cliaddr, servaddr;
    
    memset(&cliaddr, 0, sizeof(cliaddr));
    
    /* Bind our local address so that the client can send to us */
    bzero((char *) &proxyaddr, sizeof(proxyaddr));
    proxyaddr.sin_family = AF_INET; // ipv4
    proxyaddr.sin_addr.s_addr = htonl(INADDR_ANY);
    proxyaddr.sin_port = htons(PROXY_PORT);

    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET; //we're using inet
    servaddr.sin_addr.s_addr = INADDR_ANY;
    servaddr.sin_port = htons(SERVER_PORT); //set the port
    
    if (bind(sock, (struct sockaddr *) &proxyaddr, sizeof(proxyaddr)) != 0) {
        perror("binding datagram socket");
        exit(1);
    }
    
    printf("Socket has port number %d\n", ntohs(proxyaddr.sin_port));

    while (true) {
        int len;
        char message[MSGLEN];
        int n = recvfrom(sock, (char *)message, MSGLEN, MSG_WAITALL, (struct sockaddr*) &cliaddr, (unsigned int*) &len);
        if (n <= 0) {
            exit(1);
        }
        uint64_t pn;
        memcpy(&pn, message, sizeof(message));

		dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_MSEC), queue, ^{
            unsigned char value[sizeof(pn)];
            std::memcpy(value, &pn, sizeof(pn)); // binary.LittleEndian.PutUint64()
            if(sendto(sock, (const char *) value, MSGLEN, 0, (const struct sockaddr *) &servaddr, sizeof(servaddr)) < 0) {
                perror("Error sending packet");
                exit(1);
            }
        });
    }
    dispatch_main();
}
