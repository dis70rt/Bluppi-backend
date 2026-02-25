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
