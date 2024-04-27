#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <unistd.h>
#include <signal.h>
#include <string.h>
#include <time.h>
#include <malloc.h>

pthread_t *threads;
int num_threads;
void* cur_mem_ptr;

// For trace data
FILE *trace_file;
char trace_line[100];
int state;
char* workload_name;

// For cpu
char cgroup_fullpath[240] = "/sys/fs/cgroup";
// For memory
size_t prev_mem_usage;

int read_aws(char name[]) {
    char command[60];
    char result[20];
    sprintf(command, "aws s3 cp s3://mlintra/%s.state -", name);

    FILE *cmd_fp = popen(command, "r");
    if (cmd_fp == NULL) {
        fprintf(stderr, "ERROR: Failed to run aws read command\n");
        exit(1);
    }

    if (fgets(result, sizeof(result), cmd_fp) == NULL) {
        fprintf(stderr, "ERROR: Failed to get command result\n");
        exit(1);    
    }
    pclose(cmd_fp);

    int num = atoi(result);
    printf("Read AWS State: %d\n", num);
    fflush(stdout);
    return num;
}

void write_aws(int state, char name[]) {
    char command[85];
    sprintf(command, "echo %d | aws s3 cp - s3://mlintra/%s.state", state, name);

    if (system(command) < 0) {
        fprintf(stderr, "ERROR: Failed to run aws write command\n");
        exit(1); 
    }
    printf("Write AWS State: %d\n", state);
    fflush(stdout);
}

void add_memory_usage(long mem_bytes) {
    if(mem_bytes == 0) {
        return;
    }
    clock_t c1, c2;
    c1 = clock();
    // printf("Create memory usage: Todo: %ld. Need Delete?: %d\n", mem_bytes, ((mem_bytes < 2*sizeof(size_t*)) && (mem_bytes != 0)));
    // fflush(stdout);
    // Too small value to assign, or... Negtive value
    while((mem_bytes < (long)(2*sizeof(size_t*))) && (mem_bytes != 0)) {
        if(cur_mem_ptr == NULL) {
            fprintf(stderr, "Memory calculation error!!\n");
            exit(1);
        }
        char* cur_mem_char_ptr = (char*)cur_mem_ptr;
        size_t* size_ptr = (size_t*)(cur_mem_char_ptr + sizeof(size_t*));
        size_t cur_mem_bytes = *size_ptr;

        size_t* prev_mem_ptr_ptr = (size_t*)cur_mem_ptr;
        void* prev_mem_ptr = (void*)(*prev_mem_ptr_ptr);

        free(cur_mem_ptr);

        cur_mem_ptr = prev_mem_ptr;
        mem_bytes += cur_mem_bytes;

        printf("Deleted Memory chunk size: %ld\n", cur_mem_bytes);
        fflush(stdout);
    }

    if(mem_bytes > 0) {
        // Create Memory usage... From here on mem_bytes can convert to size_t
        void* new_mem_ptr = malloc((size_t)mem_bytes);
        if (new_mem_ptr == NULL) {
            fprintf(stderr, "Memory allocation filed!!\n");
            exit(1);
        }
        memset(new_mem_ptr, 0, (size_t)mem_bytes);
        
        // Store prev address
        size_t* prev_mem_ptr_ptr = (size_t*)new_mem_ptr;
        *prev_mem_ptr_ptr = (size_t)cur_mem_ptr;
        // Store current memory chunk size at next position
        char* new_mem_char_ptr = (char*)new_mem_ptr;
        size_t* size_ptr = (size_t*)(new_mem_char_ptr + sizeof(size_t*));
        *size_ptr = (size_t)mem_bytes;

        // Move current pointer
        cur_mem_ptr = new_mem_ptr;
        c2 = clock();
        printf("Create memory usage finished. Created Memory chunk size: %ld, Time: %f\n", *size_ptr, ((double) (c2 - c1)) / CLOCKS_PER_SEC);
        fflush(stdout);
    } else {
        c2 = clock();
        printf("Create memory usage finished. %ld Memory not created, Time: %f\n", mem_bytes, ((double) (c2 - c1)) / CLOCKS_PER_SEC);
        fflush(stdout);
    }
}

void create_cpu_usage(float cpu_cores) {
    // clock_t c1, c2;
    // c1 = clock();
    int cpu_int = (int)(cpu_cores * 100000);
    FILE* cgroup_file = fopen(cgroup_fullpath, "w");
    if (cgroup_file == NULL) {
        fprintf(stderr, "ERROR: Failed to open cgroups file!\n");
        exit(1);
    }

    fprintf(cgroup_file, "%d", cpu_int);

    fclose(cgroup_file);

    // c2 = clock();
    // printf("Create cpu usage finished. Time: %f\n", ((double) (c2 - c1)) / CLOCKS_PER_SEC);
    // fflush(stdout);
}

void timeout_handler(int signal) {
    // Process the file
    if (!fgets(trace_line, sizeof(trace_line), trace_file) || strlen(trace_line) == 0) {
        // Reach the end, set file pointer to beginning and redo this round.
        fseek(trace_file, 0, SEEK_SET);
        printf("End reading the trace file, start from beginning\n");
        fflush(stdout);
        state = 0;
        timeout_handler(signal);
        return;
    }
 
    int time_sec;
    float cpu;
    size_t memory;
    sscanf(trace_line, "%d %f %lu", &time_sec, &cpu, &memory);

    //Next round begin signal
    alarm(time_sec);

    // Start round
    printf("Start round, time:%d, cpu %f, memory %lu\n", time_sec, cpu, memory);
    fflush(stdout);
    
    create_cpu_usage(cpu);

    if(memory > prev_mem_usage) {
        add_memory_usage((long)(memory - prev_mem_usage));
    } else {
        add_memory_usage(-(long)(prev_mem_usage - memory));
    }

    prev_mem_usage = memory;

    state = state + 1;
    write_aws(state, workload_name);
}

void *thread_function(void *arg) {
    while (1) {}
    return NULL;
}

void get_cgroup_position() {
    // Get CPU cgroup info
    FILE* cgroup_pos_file = fopen("/proc/self/cgroup", "r");
    if (cgroup_pos_file == NULL) {
        fprintf(stderr, "Error opening cgroups info file.\n");
        return 1;
    }

    char cgroup_position[210];
    if (fgets(cgroup_position, sizeof(cgroup_position), cgroup_pos_file) == NULL) {
        fprintf(stderr, "Error reading cgroups info.\n");
        return 1;
    }
    fclose(cgroup_pos_file);
    int endpos = strlen(cgroup_position)-1;
    if(cgroup_position[endpos] == '\n') {
        // printf("Deleted new line at the end!!!\n");
        // fflush(stdout);
        cgroup_position[endpos] = '\0';
    }
    
    char *cgroup_position_start = strchr(cgroup_position, '/');
    strcat(cgroup_fullpath, cgroup_position_start);
    strcat(cgroup_fullpath, "/cpu.max");
    // printf("Cgroup position get: %s\n", cgroup_fullpath);
    // fflush(stdout);
}

int main(int argc, char *argv[]) {

    if(argc != 2){
        fprintf(stderr, "Usage: ./stress_generator workload_id\n");
        exit(1);
    }
    workload_name = argv[1];

    state = read_aws(workload_name);
    printf("Get sync state: %d\n", state);
    fflush(stdout);

    // Open trace file
    trace_file = fopen("trace.data", "r");
    if (trace_file == NULL) {
        fprintf(stderr, "Error opening trace file.\n");
        exit(1);
    }

    int cur_lines = 0;
    while(cur_lines < state) {
        fgets(trace_line, sizeof(trace_line), trace_file);
        cur_lines = cur_lines + 1;
    }

    // Bind time handler
    signal(SIGALRM, timeout_handler);
    
    get_cgroup_position();

    // Create threads for CPU
    num_threads = (int)sysconf(_SC_NPROCESSORS_ONLN);

    threads = (pthread_t *)malloc(num_threads * sizeof(pthread_t));
    if (threads == NULL) {
        fprintf(stderr, "ERROR: Thread memory allocation failed\n");
        return 1;
    }

    // Set memory using mmap. Ignore small memory assigns at beginning...
    prev_mem_usage  = 0;
    cur_mem_ptr = NULL;

    if (mallopt(M_MMAP_THRESHOLD, 0) != 1) {
        fprintf(stderr, "Error: Mallopt failed\n");
        return 1;
    }

    for (int i = 0; i < num_threads; ++i) {
        if (pthread_create(&threads[i], NULL, thread_function, NULL) != 0) {
            fprintf(stderr, "Error: creating thread %d\n", i);
            return 1;
        }
    }

    // Start the rounds
    timeout_handler(0);

    // sub threads will never terminate...
    for (int i = 0; i < num_threads; ++i) {
        pthread_join(threads[i], NULL);
    }
    fprintf(stderr, "Never executed here\n");

    free(threads);
    free(cur_mem_ptr);

    return 1;
}
