# 암호화폐 거래소 매칭 엔진 조사 및 PoC 구현 가이드

## 개요

매칭 엔진은 모든 암호화폐 거래소의 **핵심 구성 요소**로, 매수 및 매도 주문을 연결하여 거래를 촉진하는 역할을 합니다. 이 포괄적인 조사는 중앙화된 거래소부터 탈중앙화 금융(DeFi) 프로토콜에 이르기까지 다양한 유형의 매칭 엔진을 다루며, 개념 증명(PoC) 구축을 위한 사용 가능한 오픈 소스 구현도 함께 살펴봅니다.

## 매칭 엔진의 종류

### 중앙화된 거래소 매칭 엔진

**전통적인 오더북 모델**

중앙화된 거래소는 전통적인 오더북 원칙에 따라 작동하는 정교한 매칭 엔진을 사용합니다[^1][^2]. 이러한 시스템은 다음과 같은 특징을 가집니다:

- **밀리초 단위로 주문 처리**, 일부 엔진은 초당 20,000개 이상의 주문을 처리할 수 있음[^3]
- **가격-시간 우선순위 알고리즘** (FIFO - 선입선출) 준수[^1][^4]
- **시장가, 지정가, 손절매 등 다양한 주문 유형** 및 고급 주문 지원[^1][^5]
- **가격 수준별로 구성된 별도의 매수 및 매도 오더북** 유지[^6]

**주요 사용 알고리즘:**

- **FIFO (선입선출)**: 동일한 가격 수준에서 시간 우선순위에 따라 주문 체결[^1][^4]
- **비례 배분(Pro-Rata)**: 가격이 동일할 때 거래량이 많은 주문에 우선순위 부여[^1]
- **시간 가중 비례 배분(Time-weighted Pro-Rata)**: 가격과 시간 요소를 결합하여 더 나은 가격의 주문에 우선순위 부여[^1]

**성능 특징:**

최신 중앙화된 매칭 엔진은 놀라운 성능 지표를 달성합니다[^3][^7]:

- **처리 속도**: 거래당 0.05마이크로초
- **처리량**: 초당 20,000개 이상의 주문
- **지연 시간**: 밀리초 미만의 주문 매칭
- **거래 쌍**: 100개 이상의 동시 거래 쌍 지원


### 탈중앙화 금융(DeFi) 및 탈중앙화 매칭

**자동화된 시장 조성자(AMM) 대 오더북**

DeFi는 주문 매칭에 다른 접근 방식을 도입합니다[^8][^9]:

**AMM 기반 시스템:**

- **자산 비율에 따라 가격이 결정되는 유동성 풀** 활용[^8]
- **매칭 상대방 없이 지속적인 유동성** 제공[^8]
- **예시**: 유니스왑(Uniswap), 스시스왑(SushiSwap), 팬케이크스왑(PancakeSwap)

**오더북 기반 탈중앙화 거래소(DEX):**

- **온체인 또는 하이브리드 모델에서 전통적인 오더북 매칭** 운영[^8][^9]
- **오더북 깊이를 만들기 위해 사용자의 적극적인 참여** 필요[^8]
- **매수-매도 호가 상호작용을 통해 더 정확한 가격 발견** 제공[^8]

**하이브리드 모델:**

Orderly Network와 같은 프로젝트는 중앙화된 성능과 탈중앙화된 투명성을 결합합니다[^10]:

- **온체인 결제를 통한 빠른 오프체인 매칭**
- **특수 시퀀싱을 통한 MEV 보호**
- **낮은 지연 시간을 유지하면서 자체 수탁**


### MEV 보호 및 공정한 주문

**일괄 경매 시스템**

최신 매칭 엔진은 혁신적인 메커니즘을 통해 MEV(최대 추출 가능 가치) 보호 기능을 통합합니다[^11][^12]:

**빈번한 일괄 경매(FBA):**

- **동시 실행 전에 지정된 기간 동안 주문 수집**[^12]
- **선행 매매를 방지하기 위해 단일 청산 가격으로 모든 거래 실행**[^12]
- **가격 슬리피지 및 조작 기회 감소**[^12]

**MEV 보호 시스템의 예:**

- **세이 네트워크(Sei Network)**: 선행 매매 방지를 위해 일괄 경매와 함께 기본 주문 매칭 사용[^11][^13]
- **Orderly Network**: 빠른 시퀀싱을 갖춘 하이브리드 오더북 모델 구현[^10]
- **CoinToss Anti-Sandwich**: 블록 내 거래 순서 제거[^14]


## 오픈 소스 구현

### 고성능 솔루션

**1. ArjunVachhani/order-matcher (C#)**[^5]

- **성능**: AWS c6a.xlarge에서 초당 1백만 개 주문 처리
- **기능**: 지정가, 시장가, 손절매, 아이스버그 주문 지원
- **고급 옵션**: IOC, FOK, GTD, 자가 거래 방지
- **직렬화**: JSON보다 15배 빠른 맞춤형 직렬화기
- **라이선스**: MIT

**2. Dingir Exchange (Rust)**[^15]

- **아키텍처**: 완전 비동기, 단일 스레드, 메모리 기반
- **성능**: 수천 TPS
- **기술 스택**: Rust, GRPC, Tokio/Hyper/Tonic
- **지속성**: 추가 작업 로그 + Redis와 유사한 포크 앤 세이브 방식
- **영감**: Redis 및 Viabtc 거래소 아키텍처

**3. Open Outcry (Go + PostgreSQL)**[^16]

- **철학**: 매칭뿐만 아니라 전체 거래 주기의 성능
- **아키텍처**: 거래 로직을 위한 최적화된 PostgreSQL 프로시저
- **기능**: 다중 자산 지원, ACID 속성, 무중단
- **배포**: 쿠버네티스와 통합된 클라우드 네이티브
- **라이선스**: AGPL-3.0


### 특수 및 실험적 솔루션

**4. CoinTossX (Java)**[^17]

- **초점**: 저지연, 고처리량 학술 연구
- **기술**: 전송을 위한 Aeron Media Driver가 포함된 Java
- **프로토콜**: UDP SBE 메시지 프로토콜
- **배포**: 데스크톱 테스트에서 클라우드 규모의 Azure 배포까지
- **인터페이스**: Julia 및 Python 언어 지원

**5. 0x5487/matching-engine (Go)**[^18]

- **기능**: 시장가, 지정가, IOC, 포스트 온리, FOK 주문
- **아키텍처**: 고속을 위한 전체 인메모리 방식
- **다중 시장**: 여러 거래 쌍 동시 지원
- **통합**: 거래 게시를 위한 메시지 큐 통합

**6. OpenCEX (Python/Django)**[^19]

- **범위**: 완전한 암호화폐 거래소 플랫폼
- **기능**: 수탁 지갑, KYC/KYT, 전문 거래 인터페이스
- **지원 자산**: BTC, ETH, BNB, TRX, USDT (다중 네트워크)
- **배포**: 포괄적인 문서를 갖춘 Docker 기반


### 실험적인 MEV 방지 솔루션

**7. Shifu 매칭 엔진**[^14]

- **혁신**: 오더북과 풀 유동성 결합
- **MEV 보호**: 블록 내 거래 순서 제거
- **가격 발견**: 수요/공급 및 풀 유동성에 의해 결정
- **대상**: 샌드위치 공격 및 선행 매매 해결


## 기술 아키텍처 고려 사항

### 프로그래밍 언어 및 성능

**성능 요구 사항별 언어 선택:**

1. **초고성능**: C++, Rust
    - **장점**: 메모리 제어, 제로 코스트 추상화[^15][^20]
    - **사용 사례**: 기관 등급 거래소, 고빈도 거래(HFT) 시스템
2. **생산성을 갖춘 고성능**: Go, Java
    - **장점**: 빠른 개발 속도와 우수한 성능[^16][^17]
    - **사용 사례**: 중급 거래소, 신속한 프로토타이핑
3. **신속한 개발**: Python, C#
    - **장점**: 풍부한 생태계, 빠른 시장 출시 시간[^5][^19]
    - **사용 사례**: PoC 개발, 소규모 거래소

### 아키텍처 패턴

**메모리 대 데이터베이스 트레이드오프**[^21]

**인메모리 시스템:**

- **장점**: 마이크로초 수준의 지연 시간, 높은 처리량
- **단점**: 복잡한 상태 관리, 충돌 복구
- **최적**: 고빈도 거래, 최대 성능

**데이터베이스 중심 시스템:**

- **장점**: ACID 속성, 쉬운 복구, 데이터 일관성
- **단점**: 높은 지연 시간, 낮은 최대 처리량
- **최적**: 전통적인 거래소, 규제 준수

**하이브리드 접근 방식:**

- **메모리 기반 매칭**과 데이터베이스 지속성 결합
- **충돌 복구를 위한 추가 전용 로그** 사용[^15]
- **성능을 위한 비동기 데이터베이스 쓰기** 구현


## PoC 구현 권장 사항

### 학습 및 프로토타이핑용

**최적의 시작점:**

1. **ArjunVachhani/order-matcher**: 훌륭한 문서, 최신 기능, 활발한 유지보수[^5]
2. **0x5487/matching-engine**: 간단한 Go 구현, 이해하기 쉬움[^18]
3. **Open Outcry**: 독특한 데이터베이스 중심 접근 방식, 프로덕션 준비 개념[^16]

### 프로덕션 고려 사항

**확장성 요소:**

- **시장 세분화를 통한 수평적 확장**[^22]
- **다양한 구성 요소를 위한 마이크로서비스 아키텍처**[^23]
- **로드 밸런싱 및 장애 조치 메커니즘**[^24]
- **실시간 시장 데이터 배포**[^6]

**성능 최적화:**

- **오더북 저장을 위한 레드-블랙 트리**[^21]
- **비동기 작업을 위한 메시지 큐**[^21]
- **메모리 관리 최적화**[^21]
- **네트워크 프로토콜 선택** (TCP 대 UDP)[^17]


### 구현 전략

**1단계: 기본 PoC**

1. 간단한 오픈 소스 구현 선택[^5][^18]
2. 기본 주문 유형(시장가, 지정가) 설정
3. 간단한 가격-시간 우선순위 매칭 구현
4. 기본 오더북 시각화 추가

**2단계: 향상된 기능**

1. 고급 주문 유형(손절매, IOC, FOK) 추가
2. 시장 데이터 피드 구현
3. 기본 위험 관리 추가
4. 간단한 웹 인터페이스 생성

**3단계: 고급 개념**

1. MEV 보호 메커니즘 탐색[^11][^12]
2. 일괄 경매 기능 구현
3. 다중 시장 지원 추가
4. 하이브리드 중앙화/탈중앙화 모델 고려

암호화폐 거래소 매칭 엔진 환경은 전통적인 고성능 중앙화 시스템부터 MEV 보호 기능을 갖춘 혁신적인 DeFi 솔루션에 이르기까지 다양한 접근 방식을 제공합니다. 오픈 소스 구현은 학술 연구 프로젝트부터 프로덕션 준비 시스템에 이르기까지 PoC 개발을 위한 훌륭한 시작점을 제공합니다. 선택은 성능, 기능 및 아키텍처 철학에 대한 특정 요구 사항에 따라 달라집니다.

<div style="text-align: center">⁂</div>

[^1]: https://b2broker.com/news/what-is-cryptocurrency-matching-engine/
[^2]: https://paybis.com/blog/glossary/matching-engine/
[^3]: https://www.bitdeal.net/cryptocurrency-matching-engine
[^4]: https://academy.binance.com/en/articles/understanding-matching-engines-in-trading
[^5]: https://github.com/ArjunVachhani/order-matcher
[^6]: https://liquidity-provider.com/articles/order-matching-engine-the-heart-of-a-crypto-exchange/
[^7]: https://blog.kaiko.com/bitstamps-new-matching-engine-how-nasdaq-has-improved-trading-frequency-and-order-book-efficiency-14e8825dc138?gi=cdd499f68365
[^8]: https://www.lbank.com/questions/armqu51742348867
[^9]: https://www.nadcab.com/blog/order-matching-in-decentralized-exchange-development
[^10]: https://www.xcritical.in/blog/crypto-matching-engine-what-is-and-how-does-it-work/
[^11]: https://collective.flashbots.net/t/mev-share-programmably-private-orderflow-to-share-mev-with-users/1264
[^12]: https://github.com/quantyle/matching-engine
[^13]: https://mvpworkshop.co/order-book-vs-amm-which-one-will-win/
[^14]: https://github.com/CoinFuMasterShifu/shifu-matching
[^15]: https://github.com/fluidex/dingir-exchange
[^16]: https://github.com/tolyo/open-outcry
[^17]: https://arxiv.org/abs/2102.10925
[^18]: https://github.com/0x5487/matching-engine
[^19]: https://github.com/Polygant/OpenCEX
[^20]: https://github.com/pgellert/matching-engine
[^21]: https://blog.valensas.com/matching-engine-design-for-high-throughput-consistency-and-availability-3b1134613d35?gi=54592c5789ea
[^22]: https://github.com/ngquyduc/matching-engine-go
[^23]: https://www.antiersolutions.com/blogs/understanding-the-role-of-order-matching-engine-in-a-centralized-crypto-exchange/
[^24]: https://finchtrade.com/glossary/matching-engine-architecture
[^25]: https://www.kanikhetook.tv/understanding-matching-engines-in-buying-and-2/
[^26]: https://www.soft-fx.com/technologies/matching-engine/
[^27]: https://www.alwin.io/order-matching-algorithms-in-crypto-exchange
[^28]: https://www.nadcab.com/blog/order-matching-affect-trading-on-dex
[^29]: https://databento.com/microstructure/matching-engine
[^30]: https://euralex.org/wp-content/themes/euralex/proceedings/Euralex 1998 Part 1/Archibald MICHIELS The DEFI Matcher.pdf
[^31]: https://cointelegraph.com/innovation-circle/centralized-vs-decentralized-orders-matching-on-dexs
[^32]: https://www.openware.com/projects/finance/matching-engine
[^33]: https://github.com/victorlaiyeeteng/crypto-exchange
[^34]: https://dhyeymavani.com/project/matchingengine/
[^35]: https://chronicle.software/wp-content/uploads/2023/01/Chronicle-Matching-Engine.pdf
[^36]: https://github.com/marekpinto/exchangematchingengine
[^37]: https://github.com/wailo/orderbook-matching-engine
[^38]: https://www.sciencedirect.com/science/article/pii/S2352711022000875
[^39]: https://databento.com/blog/matching-engines-guide
[^40]: https://www.libhunt.com/topic/matching-engine
[^41]: https://www.goglides.dev/bellabardot/guide-to-choose-the-right-tech-stack-for-crypto-exchange-development-220b
[^42]: https://en.wikipedia.org/wiki/Pattern_matching
[^43]: https://softwareengineering.stackexchange.com/questions/148045/programming-language-with-pattern-matching-in-trees
[^44]: https://www.shoal.gg/p/mev-protection-dex-and-aggregator
[^45]: https://www.odaily.news/en/post/5187715
[^46]: https://mirror.xyz/3vlabs.eth/NqNhcS38WVbruwjzgzpQJMBOT2wrXqi8b9S20ZPV_eM?collectors=true
[^47]: https://orderly.network/blog/mev-olution-ensuring-fairness-with-orderly-s-unique-orderbook-model/
[^48]: https://www.gate.io/post/status/3817841
[^49]: https://tangem.com/en/glossary/batch-auctions/
[^50]: https://www.youtube.com/watch?v=fCP88sHy8iA
[^51]: https://www.osiztechnologies.com/blog/choosing-right-technology-stack-to-build-crypto-exchange-platform
[^52]: https://github.com/topics/matching-engine?l=javascript
[^53]: https://github.com/AnshuJalan/tezos-nft-batch-auction
[^54]: https://www.bitget.com/glossary/matching-engine
