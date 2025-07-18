# PoC Matching Engine Specification

## 1. Introduction

This document outlines the architectural vision for a Proof-of-Concept (PoC) matching engine. The primary goal of this project is not to develop a production-ready system, but rather to serve as an exploration ground for architecting complex systems under various challenging requirements. It aims to foster deeper understanding and uncover novel viewpoints that can be applied to future work, particularly in domains requiring high scalability and resilience, such as Web3 applications.

The implementation will leverage Go, chosen for its simplicity, productivity, and emphasis on clear code. In a system as critical and complex as a matching engine, code clarity is paramount to minimizing errors, as even minor confusion can lead to significant issues. The matching engine component must inherently be error-free within the broader system.

## 2. Core Architectural Principles & Ideas

### 2.1. Functional Scope

The matching engine server will exclusively focus on the core logic of matching orders. It will not incorporate business logic unrelated to order matching. However, it will support a flexible "afterOrder" handler function. This mechanism allows for the concurrent execution of additional processes, such as updating state changes to a PostgreSQL database for persistence, without blocking the core matching process. Each order will include an `orderer_id` to uniquely identify the participant.

### 2.2. Transaction Sequencing & Concurrency

A fundamental constraint of order matching is its inherent sequential nature. Each incoming transaction (order) often has a dependency on the state resulting from previous transactions. Altering the order of processing would directly change the outcome of the matching engine. Consequently, the core matching process cannot be truly parallelized in the traditional sense, as transaction order must be strictly preserved.

However, exploration will be made into specific scenarios where parallel execution might be feasible. For instance, certain "cancel" orders that do not require mutex locks on critical order book ranges could potentially be processed in parallel. Further consideration will be given to the hypothesis that in high-frequency trading environments, liquidity tends to be concentrated within small price ranges. If this holds true, it might be possible to parallelize the processing of orders that fall significantly outside this "hot" price range, where "real trade executions are still limited." The viability of this approach will need careful evaluation to ensure it doesn't introduce undue complexity without providing tangible performance benefits.

### 2.3. System Configuration

The matching engine will support configurable parameters, such as a `minimum_tick` size, which defines the smallest allowable price increment.

### 2.4. Scalability Approach

While the initial design aims to avoid deploying multiple servers for different assets, a mechanism for splitting server responsibilities will be considered. This will allow for the horizontal scaling of the matching engine by dedicating separate instances to assets exhibiting consistently heavy order flow, thus preventing bottlenecks.

### 2.5. Order Types

Beyond standard market, limit, and stop-loss orders, the engine will explore support for more specialized order types crucial for diverse trading strategies:

*   **Post-Only Orders:** Designed to add liquidity to the order book. A Post-Only order will only execute if it does not immediately match with an existing order. If it would "cross the spread" and act as a market taker, it will be rejected. This type is valuable for liquidity providers aiming to earn maker rebates by avoiding taker fees. [1]
*   **All-or-None (AON):** An order that must be executed in its entirety, or not at all.
*   **Fill-or-Kill (FOK):** An order that must be immediately executed in its entirety, or be cancelled.
*   **Immediate-or-Cancel (IOC):** An order that must be immediately executed (partially or fully), with any unexecuted portion cancelled. [1]

### 2.6. Testing Framework

A robust testing framework is critical for validating the matching engine's performance and correctness. This framework will:

*   Enable benchmarking of processing speed against existing open-source matching engine solutions.
*   Facilitate comprehensive testing under various special scenarios. This will involve exploring diverse trading scenarios and generating specific test data sets to ensure the project's capabilities are thoroughly verified.

### 2.7. Insights from Blockchain Architectures

The operational characteristics of blockchain platforms, particularly Ethereum, offer valuable insights for a matching engine. The reasons platforms like OKX have opted to build their own blockchains (e.g., from Ethereum's codebase) resonate with matching engine requirements:

*   **Wallet and Authentication Systems:** Secure user identification and asset management.
*   **Sequential Transaction Processing:** The inherent ordered nature of blockchain transactions aligns with the need for strict transaction sequencing in a matching engine.
*   **24/7 No-Downtime Operation:** Blockchain networks are designed for continuous availability, a critical requirement for cryptocurrency exchanges.
*   **Guaranteed Transaction Propagation:** Valid transactions, once broadcast, are designed to propagate and eventually be confirmed, preventing "missing" user transactions.

### 2.8. In-Memory State Management (Inspired by LMAX)

Drawing inspiration from architectures like LMAX, an in-memory state management approach will be prioritized. This offers significant speed advantages, which are paramount for high-frequency trading environments. For 24/7 crypto exchanges, any system downtime is highly problematic, even if sacrificing some speed for persistence. Therefore, favoring in-memory operations while carefully managing persistence (e.g., via the `afterOrder` handler) aligns with this need. [2]

Furthermore, the impact of network latency, particularly when physical servers are distant from users, necessitates careful consideration of transaction size. Efforts will be made to minimize the size of each transaction and reduce the number of parameters in unpacked transactions to optimize network transfer efficiency. [2]

## References

[1] https://b2broker.com/news/what-is-cryptocurrency-matching-engine/
[2] https://martinfowler.com/articles/lmax.html#KeepingItAllInMemory
