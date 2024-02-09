// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly
#define _OPEN_THREADS
#include <pthread.h>
#include <stdio.h>

int i = 0;
pthread_t tid[2];
pthread_mutex_t lock;

int j;
// Note the return type: void*
void* incrementingThreadFunction(){
    // TODO: increment i 1_000_000 times
    pthread_mutex_lock(&lock);
    for (j = 0; j < 1000000; j++) {
        i= i+1;
    }
    pthread_mutex_unlock(&lock);
    return NULL;
}

int k;
void* decrementingThreadFunction(){
    pthread_mutex_lock(&lock);
    // TODO: decrement i 1_000_000 times
    for (k = 0; k < 1000001; k++) {
        i= i-1;
    }
    pthread_mutex_unlock(&lock);
    return NULL;
}


int main(){
    pthread_t ptid; 
    pthread_t ptid2; 
        
    if (pthread_mutex_init(&lock, NULL) != 0) {
        printf("\n mutex init failed\n");
        return 1;
    }

    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?
    
    
    // Creating a new thread 
    pthread_create(&ptid, NULL, incrementingThreadFunction, NULL); 
    pthread_create(&ptid2, NULL, decrementingThreadFunction, NULL); 


    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`  p
    pthread_join(ptid, NULL);
    pthread_join(ptid2, NULL);
    pthread_mutex_destroy(&lock);

    printf("The magic number is: %d\n", i);
    return 0;
}