#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <unistd.h>
#include <signal.h>
#include <string.h>

pthread_t *threads;
int num_threads;
void* target_mem;

void timeout_handler(int signal) {
    printf("Received timeout signal. Destroy all..\n");
    // Quit the whole program...

    free(threads);
    free(target_mem);
    exit(0);
}

void *thread_function(void *arg) {
    while (1) {}
    return NULL;
}

int main(int argc, char *argv[]) {

    if (argc != 3) {
        fprintf(stderr, "Usage: %s <seconds> <size_memory_bytes>\n", argv[0]);
        return 1;
    }

    num_threads = (int)sysconf(_SC_NPROCESSORS_ONLN);

    int time_sec = atoi(argv[1]);
    signal(SIGALRM, timeout_handler);
    alarm(time_sec);

    char *endptr;
    size_t memory_bytes = strtoul(argv[2], &endptr, 10);
    if (*endptr != '\0') {
        printf("Conversion failed: %s is not a valid number.\n", argv[2]);
        return 1;
    }

    // Create threads for dead loop CPU usage...
    threads = (pthread_t *)malloc(num_threads * sizeof(pthread_t));
    if (threads == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        return 1;
    }
    for (int i = 0; i < num_threads; ++i) {
        if (pthread_create(&threads[i], NULL, thread_function, NULL) != 0) {
            fprintf(stderr, "Error creating thread %d\n", i);
            return 1;
        }
    }
    printf("%d All threads created.\n", num_threads);


    // Create Memory usage...
    memory_bytes = memory_bytes - num_threads*sizeof(pthread_t*);
    target_mem = malloc(memory_bytes);
    if (target_mem == NULL) {
        fprintf(stderr, "Memory allocation filed!!\n");
        exit(1);
    }
    memset(target_mem, 0, memory_bytes);
    printf("Created memory %ld bytes.\n", memory_bytes);


    // sub threads will never terminate...
    for (int i = 0; i < num_threads; ++i) {
        pthread_join(threads[i], NULL);
    }
    fprintf(stderr, "Never executed here\n");
    return 1;
}
