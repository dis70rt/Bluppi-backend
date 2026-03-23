# Bluppi Distributed Audio Synchronization

**Low-Echo / Low-Phasing Playback Architecture**

---

## 1. Perceptual & Acoustic Constraints

Human perception defines hard failure limits for distributed playback.

If inter-device delay exceeds ~40 ms, the auditory system no longer fuses signals:

* Sound becomes a **distinct echo**
* Unified amplification illusion collapses

Reference:

* Haas / Precedence Effect
  [https://www.wikiaudio.org/haas-effect/](https://www.wikiaudio.org/haas-effect/)

---

### Comb Filtering & Phasing Artifacts

Complex music contains broadband frequency content. Even tiny timing errors produce:

* Phase cancellations
* “Hollow” or “swishy” sound
* Artificial spatial coloration

Avoiding comb filtering requires **extremely tight alignment**.

Reference (IDMS precision & Wi-Fi synchronization):

* OpenWiFiSync Research
  [https://www.researchgate.net/publication/384887259_OpenWiFiSync_Open_Source_Implementation_of_a_Clock_Synchronization_Algorithm_using_Wi-Fi](https://www.researchgate.net/publication/384887259_OpenWiFiSync_Open_Source_Implementation_of_a_Clock_Synchronization_Algorithm_using_Wi-Fi)

---

### Empirical Delay Thresholds

| Acoustic Phenomenon | Time Delay Range | Perceptual Result                 | Engineering Implication              |
| ------------------- | ---------------- | --------------------------------- | ------------------------------------ |
| **Comb Filtering**  | 0.1 – 15 ms      | Hollow / phase coloration         | Requires microsecond-scale stability |
| **Haas Fusion**     | 15 – 35 ms       | Single fused image, widened field | Perceptually usable, not coherent    |
| **Discrete Echo**   | > 40 ms          | Audible repetition                | System failure                       |

Reference (psychoacoustics & delay perception):

* Lund University Thesis
  [https://lup.lub.lu.se/luur/download?func=downloadFile&recordOId=8052964&fileOId=8053165](https://lup.lub.lu.se/luur/download?func=downloadFile&recordOId=8052964&fileOId=8053165)

---

## 2. Clock Instability Reality

Every client device contains imperfect oscillators.

Even identical hardware exhibits:

* Frequency deviation (clock skew)
* Absolute time divergence (clock offset)

Clock model:

$$
C_i(t) = \alpha_i t + \theta_i
$$

Where:

* ( $\alpha_i$ ) → clock skew
* ( $\theta_i$ ) → clock offset

References:

* Clock Drift & Synchronization Limits
  [https://arxiv.org/pdf/1607.03830](https://arxiv.org/pdf/1607.03830)

* Offset vs Skew Explanation
  [https://www.baeldung.com/cs/clock-offset-skew-difference](https://www.baeldung.com/cs/clock-offset-skew-difference)

* Wireless Clock Sync Accuracy Limits
  [https://arxiv.org/pdf/2212.07138](https://arxiv.org/pdf/2212.07138)

---

## 3. System Design Principles

Blupii avoids network-driven playback decisions.

Key rules:

1. Playback always scheduled in the future
2. Clients estimate offset & skew locally
3. Network delay never directly controls audio timing
4. Playback rate absorbs residual error

---

## 4. Initialization Phase (Burst Sampling)

At session start:

* Client transmits **10 rapid synchronization probes**
* Objective → isolate minimum-delay path

Each probe uses a PTP-style exchange.

References:

* Precision Time Protocol (IEEE 1588 Concepts)

---

### Timestamp Exchange

Four timestamps:

* ( $t_1$ ) → client send
* ( $t_2$ ) → server receive
* ( $t_3$ ) → server transmit
* ( $t_4$ ) → client receive

Delay estimation:

$$
delay = \frac{(t_2 - t_1) + (t_4 - t_3)}{2}
$$

Offset estimation:

$$
offset = \frac{(t_2 - t_1) - (t_4 - t_3)}{2}
$$

---

## 5. Jitter Suppression (Critical Step)

Mobile networks introduce stochastic queuing delays.

Strategy:

1. Collect burst delay samples
2. Select **minimum-delay measurement**

Reasoning:

* Queuing only increases latency
* Minimum delay approximates physical propagation

Conceptually aligned with first-arrival dominance methods.

Reference:

* RMTS vs FTSP Behavior (Wireless Sensor Networks)
  [https://arxiv.org/pdf/1607.03830](https://arxiv.org/pdf/1607.03830)

---

## 6. Drift / Skew Estimation

Offsets evolve linearly under skew:

$$
offset(t) \approx (\alpha_i - 1)t + \theta_i
$$

Skew approximation:

$$
\alpha_i \approx 1 + \frac{offset(t_2) - offset(t_1)}{t_2 - t_1}
$$

Continuous refinement required.

---

## 7. Kalman Filter State Tracking (Recommended)

Clock behavior is noisy and dynamic.

State vector:

$$
x(k) =
\begin{bmatrix}
\theta(k) \
\alpha(k)
\end{bmatrix}
$$

Prediction:

$$
x(k) =
\begin{bmatrix}
1 & \tau_0 \
0 & 1
\end{bmatrix}
x(k-1) + \omega(k)
$$

Measurement:

$$
z(k) = \theta(k) + v(k)
$$

Reference:

* Kalman-Based Clock Synchronization
  [https://www.mdpi.com/1424-8220/21/13/4426/pdf](https://www.mdpi.com/1424-8220/21/13/4426/pdf)

---

## 8. Playback Scheduling Strategy

Server never issues immediate play commands.

Instead:

$$
T_{play} = T_s + \Delta
$$

Where:

* ( $\Delta$ ) = 4–5 seconds future horizon

Benefits:

* Network jitter irrelevant to start alignment
* Deterministic scheduling

---

## 9. Local Playback Time Conversion

Client mapping:

$$
t_{local} = \frac{T_{play} - \theta_i}{\alpha_i}
$$

Playback is clock-corrected, not packet-timed.

---

## 10. Drift & Phasing Control

Clock drift is permanent.

Hard corrections cause audible artifacts.
Use **slow playback rate correction**:

$$
rate_{adj} = 1 - k \cdot error
$$

Reference concepts:

* Oscillator Noise & Drift Behavior
  [https://arxiv.org/pdf/2212.07138](https://arxiv.org/pdf/2212.07138)

---

## 11. Error Threshold Handling

| Alignment Error | Action                   |
| --------------- | ------------------------ |
| < 1 ms          | Stable                   |
| 1 – 15 ms       | Increase rate correction |
| > 30 ms         | Rebuffer / reschedule    |
| > 40 ms         | Audible echo → reset     |

Derived from psychoacoustic limits:

* Haas / Echo Perception
  [https://www.wikiaudio.org/haas-effect/](https://www.wikiaudio.org/haas-effect/)

---

## 12. Physical Acoustic Limitation

Even perfect clock sync cannot remove spatial delay:

* ~3 ms per meter sound propagation

Reference:

* Psychoacoustic Delay Perception Studies
  [https://lup.lub.lu.se/luur/download?func=downloadFile&recordOId=8052964&fileOId=8053165](https://lup.lub.lu.se/luur/download?func=downloadFile&recordOId=8052964&fileOId=8053165)

---

## 13. Design Outcome

Stable distributed playback requires:

* Offset estimation
* Drift prediction
* Jitter rejection
* Rate correction

Clock sync alone is insufficient.


## Presence Gateway Architecture: Multi-Device Support & Targeted Fanout

Our presence gateway is designed to efficiently handle real-time online/offline status updates across a distributed system. 

### The Problem: Single-Channel Overwrites & Broadcast Spam
Initially, the gateway mapped a single `UserID` to a single channel. This meant if a user opened the app on their phone and a web browser simultaneously, the second connection would overwrite the first. When one device disconnected, it would close the channel for the other. 

Furthermore, listening to presence events meant listening to *every* event in the system, forcing the application to filter massive amounts of irrelevant traffic, wasting bandwidth and client CPU.

### The Solution: Targeted Subscriptions & Multi-Device Fanout
We restructured the in-memory `ConnectionManager` to support concurrent device sessions and granular event routing.

1. **Multi-Device Connection Indexing**: 
   Instead of a flat `map[string]chan`, the gateway now uses `map[string]map[string]*Connection` (`userId` -> `connectionId` -> `Connection`).
   * **Result:** A user can connect from a mobile app, a tablet, and a web browser simultaneously. Each connection gets a unique UUID, ensuring that one tab disconnecting only cleans up its own stream, leaving the others active.

2. **Targeted Watchers (Subscribed Fanout)**:
   Clients now pass an array of `TargetUserIDs` in their gRPC `SubscribePresence` request (e.g., the IDs of the people in their current room or friend list). 
   * **Result:** The manager maintains an inverted index (`map[string]map[string]struct{}`) mapping a `TargetUserID` to the set of `connectionIds` watching them. 

3. **O(1) Event Routing**:
   When the Redis PubSub listener receives an online/offline event for `User A`, it no longer broadcasts it to everyone. Instead, it looks up `User A` in the watcher index and pushes the event *only* to the specific connections that requested to watch `User A`.
   
4. **Resilient Streaming & Safe Cleanup**:
   We introduced robust channel closure handling (`case event, ok := <-conn.Chan`) to prevent zero-value spam loops. When a stream drops, the `RemoveConnection` method cleans up the exact connection ID from the user's connection map and removes it from all associated watcher target sets in O(N) time (where N is the number of targets watched by that specific connection).

---

## Presence Load Testing Report

This report summarizes real-world load-test behavior of the presence gateway using a single laptop as the load generator.

### Test Context

- Scenario: Presence connection and message-receive validation
- Objective: 100,000 active users
- Constraint: Run stopped early due to load-generator RAM pressure
- Peak reached: 51,322 active users

### Results Overview

| Run Target | Time Elapsed | Active/Target | Total Requests | Errors | Messages Rx | Connect Rate | p50 | p90 | p95 | p99 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 50 | 37s | 0 / 50 | 50 | 0 | 50 | 1.35 req/s | 1.365 ms | 1.595 ms | 1.769 ms | 3.036 ms |
| 5,000 | 332s | 0 / 5,000 | 5,000 | 0 | 5,000 | 15.06 req/s | 0.756 ms | 0.933 ms | 0.992 ms | 1.231 ms |
| 20,000 | 632s | 0 / 20,000 | 20,000 | 0 | 20,000 | 31.64 req/s | 0.713 ms | 2.490 ms | 2.636 ms | 2.783 ms |
| 100,000 target (stopped) | 308s | 51,322 / 100,000 | 51,322 | 0 | 51,322 | 166.63 req/s | 2.695 ms | 8.426 ms | 11.516 ms | 15.278 ms |

### Key Findings

- Reliability remained strong at scale: zero request errors up to 51,322 connections.
- Tail latency stayed controlled under heavy load: p99 connect latency remained near 15 ms.
- The stopping condition was load-generator resource exhaustion, not a backend correctness failure.
- Current data supports backend stability at least through the 50k concurrent-user range for this workload shape.

### What This Means (Simple Read)

- The presence system handled a large concurrency level without request failures.
- Latency increased as expected at higher scale, but remained in a healthy range.
- The test was limited by local machine resources before backend saturation was reached.
- A multi-generator run is still needed to claim a confirmed 100k-user capacity.

### Benchmark Host Profile

```text
 /\___/\       dis70rt@fedora
 )     (
=\     /=      distribution    - Fedora Linux 43 (KDE Plasma Desktop Edition)
  )   (        linux kernel    - Linux 6.19.8-200.fc43.x86_64
 /     \       cpu             - Intel Core i5-13500HX (14 cores / 20 threads)
 )     (       memory          - 15 GiB RAM, 8 GiB swap
/       \      gpus            - Intel UHD + NVIDIA RTX 4050 Mobile
\       /      terminal/shell  - kitty 0.43.1 / zsh 5.9
 \__ __/       wm              - Hyprland (Wayland)
```

### Hardware Snapshot

- CPU: 13th Gen Intel Core i5-13500HX (14 cores, 20 threads, up to 4.7 GHz)
- RAM: 15 GiB
- Swap: 8 GiB
- Storage: 476.9 GB NVMe + 931.5 GB NVMe
- GPU: Intel UHD Graphics + NVIDIA RTX 4050 Mobile
- Kernel: Linux 6.19.8-200.fc43.x86_64
- Open file limit during test: 1024

### Recommended Next Steps

1. Run a 30-60 minute soak test at 20k-40k users.
2. Re-run high-concurrency ramp with smaller steps (for example, +5k every 2-3 minutes).
3. Record backend CPU, memory, Redis/Postgres connections, and network throughput at peak.
4. Re-test after raising open-file limits to remove client-side socket constraints.

### Quick Conclusion

The presence gateway shows strong reliability and good latency behavior through 51k concurrent users in a single-machine test setup. The next milestone is validating full 100k behavior with higher client-side limits and/or distributed load generation.