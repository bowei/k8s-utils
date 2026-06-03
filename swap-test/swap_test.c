#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <time.h>
#include <sys/mman.h>
#include <sys/types.h>
#include <unistd.h>
#include <errno.h>

#define DEFAULT_SIZE_MB 1024
#define PAGE_SIZE 4096UL

static double elapsed_ms(struct timespec *start, struct timespec *end) {
    return (end->tv_sec - start->tv_sec) * 1000.0 +
           (end->tv_nsec - start->tv_nsec) / 1e6;
}

static void touch_pages(volatile char *buf, size_t size) {
    unsigned int seed = 42;
    for (size_t i = 0; i < size; i += PAGE_SIZE) {
        buf[i] = (char)rand_r(&seed);
    }
}

static volatile char sink;

static void read_pages(volatile char *buf, size_t size) {
    char acc = 0;
    for (size_t i = 0; i < size; i += PAGE_SIZE) {
        acc ^= buf[i];
    }
    sink = acc;
}

static void read_pages_random(volatile char *buf, size_t size) {
    size_t npages = size / PAGE_SIZE;
    uint32_t *order = malloc(npages * sizeof(uint32_t));
    if (!order) {
        fprintf(stderr, "malloc failed for page order array\n");
        exit(1);
    }
    for (size_t i = 0; i < npages; i++)
        order[i] = (uint32_t)i;

    /* Fisher-Yates shuffle */
    unsigned int seed = 42;
    for (size_t i = npages - 1; i > 0; i--) {
        size_t j = (size_t)rand_r(&seed) % (i + 1);
        uint32_t tmp = order[i];
        order[i] = order[j];
        order[j] = tmp;
    }

    char acc = 0;
    for (size_t i = 0; i < npages; i++)
        acc ^= buf[(size_t)order[i] * PAGE_SIZE];
    sink = acc;
    free(order);
}

int main(int argc, char *argv[]) {
    size_t size_mb = DEFAULT_SIZE_MB;
    int random_mode = 0;
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-r") == 0 || strcmp(argv[i], "--random") == 0) {
            random_mode = 1;
        } else if (strcmp(argv[i], "-s") == 0 || strcmp(argv[i], "--size") == 0) {
            if (i + 1 < argc) {
                long val = strtol(argv[++i], NULL, 10);
                if (val <= 0) {
                    fprintf(stderr, "Invalid size: %s (must be a positive integer)\n", argv[i]);
                    return 1;
                }
                size_mb = (size_t)val;
            } else {
                fprintf(stderr, "Option %s requires an argument\n", argv[i]);
                return 1;
            }
        } else {
            fprintf(stderr, "Usage: %s [-r|--random] [-s|--size <size_mb>]\n", argv[0]);
            return 1;
        }
    }

    size_t alloc_size = size_mb * 1024UL * 1024UL;

    struct timespec t0, t1;

    printf("Mode        : %s\n", random_mode ? "random (worst-case)" : "sequential");
    printf("Allocating %lu MB...\n", size_mb);
    char *buf = mmap(NULL, alloc_size, PROT_READ | PROT_WRITE,
                     MAP_ANONYMOUS | MAP_PRIVATE, -1, 0);
    if (buf == MAP_FAILED) {
        perror("mmap");
        return 1;
    }

    printf("Touching all pages...\n");
    touch_pages(buf, alloc_size);
    printf("Pages touched.\n\n");

    /* Swap out */
    printf("Swapping out via MADV_PAGEOUT...\n");
    clock_gettime(CLOCK_MONOTONIC, &t0);
    if (madvise(buf, alloc_size, MADV_PAGEOUT) != 0) {
        perror("madvise MADV_PAGEOUT");
        munmap(buf, alloc_size);
        return 1;
    }
    clock_gettime(CLOCK_MONOTONIC, &t1);

    double swap_out_ms = elapsed_ms(&t0, &t1);
    double swap_out_bw = (double)size_mb / (swap_out_ms / 1000.0);
    printf("SWAP_OUT_MS=%.2f\n", swap_out_ms);
    printf("SWAP_OUT_BW_MB_S=%.1f\n\n", swap_out_bw);

    /* Page in by reading every page */
    printf("Paging in by %s all pages...\n",
           random_mode ? "randomly faulting in" : "reading");
    clock_gettime(CLOCK_MONOTONIC, &t0);
    if (random_mode)
        read_pages_random(buf, alloc_size);
    else
        read_pages(buf, alloc_size);
    clock_gettime(CLOCK_MONOTONIC, &t1);

    double page_in_ms = elapsed_ms(&t0, &t1);
    double page_in_bw = (double)size_mb / (page_in_ms / 1000.0);
    printf("PAGE_IN_MS=%.2f\n", page_in_ms);
    printf("PAGE_IN_BW_MB_S=%.1f\n", page_in_bw);

    munmap(buf, alloc_size);
    return 0;
}
