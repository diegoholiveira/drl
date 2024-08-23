# drl - Diego's Rate Limiter

**WARNING: this was a study project and should not be used in production.**

## Description

This implementation uses the Sliding Window algorithm. How it works:

- Sliding Window Counter divides time into small intervals, in our case here it's hard coded to be 10 seconds.
- It maintains counters for each of these small intervals.
- When a request comes in, the algorithm checks how many requests have been made in the current sliding window.
