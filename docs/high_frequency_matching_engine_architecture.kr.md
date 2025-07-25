# 고빈도 매칭 엔진 최적화: 확장 가능하고 복원력 있는 아키텍처
이 문서는 다양한 최적화 전략에 대한 협력적 논의를 바탕으로 고빈도 매칭 엔진 구축을 위한 아키텍처 고려 사항, 즉 성능, 복원력, 운영 확장성에 중점을 두어 탐구합니다.

## 나의 핵심 아키텍처 아이디어
고빈도 매칭 엔진의 경우, 핵심 매칭 로직은 근본적으로 단일 스레드 프로그램이어야 한다고 생각합니다. 이는 모든 트랜잭션의 결과가 이전 상태에 결정적으로 의존하며, 이는 다시 이전 트랜잭션의 영향을 받기 때문입니다. 이러한 단일 스레드 시스템에서 로직 단계당 평균 실행 시간은 비교적 일정하며 일반적으로 정규 분포를 보입니다.

이를 감안할 때, 매칭 엔진 내의 링 버퍼는 일시적인 저항 메커니즘으로만 작동해야 한다고 제안합니다. 주요 역할은 마이크로 버스트 주문을 완화하고 시스템이 증가된 부하에 반응할 수 있는 짧은 창을 제공하는 매우 단기적인 버퍼를 제공하는 것입니다. 장기적인 오버플로 저장을 위해 설계되지 않았습니다.

오버플로 트랜잭션을 다른 내부 데이터 저장소(예: 유연한 힙 또는 캐시)로 전환하여 복잡성을 더하는 대신, 지속적인 높은 트랜잭션 볼륨을 처리하기 위한 다른 접근 방식에 집중하고 싶습니다. 저의 핵심 통찰력은 시스템의 처리 속도(초당 매칭 작업), 내부 링 버퍼의 현재 용량/깊이, 들어오는 이벤트의 비율(스태킹 속도)을 지속적으로 측정하는 것입니다. 이러한 메트릭을 기반으로 입력 속도가 처리 능력을 훨씬 초과하는 상당하고 지속적인 불균형이 있는 경우, 해결책은 더 많은 버퍼링이 아니라 선제적인 수직 확장입니다. 이는 처리를 인수하거나 증강하기 위해 더 강력한 계산 용량을 가진 새 서버의 배포를 트리거하는 것을 의미합니다.

이 전략이 특히 강력한 CI/CD에 적합하고 전환 중 트랜잭션 손실을 방지하려면 이벤트 시퀀싱 및 큐잉이 매칭 엔진 자체 외부에서 발생해야 한다고 결론지었습니다. 이 외부 구성 요소는 내구성, 엄격한 순서 지정을 보장하고 전체 시스템의 원활한 확장 및 복원력을 가능하게 하는 데 중요해집니다.

### 핵심: 단일 스레드 매칭 엔진(상세 설명)
핵심 아이디어에서述べたように, 핵심 매칭 로직을 단일 스레드 프로세스로 구현하는 것은 고빈도 거래 시스템에 널리 채택되고 적극 권장되는 접근 방식입니다.

*   **결정론 및 엄격한 순서 지정:** 금융 시장은 주문 실행에 절대적인 정밀도를 요구합니다. 단일 스레드 코어는 모든 트랜잭션이 순차적으로 처리되도록 보장하여 올바른 시장 상태, 정확한 가격 발견을 보장하고 작업 순서가 가장 중요한 경쟁 조건을 방지합니다. 이는 매칭 알고리즘 자체 내에서 복잡한 동시성 문제를 본질적으로 해결합니다.
*   **단순성 및 예측 가능성:** 핵심 로직 내에서 잠금, 뮤텍스 및 복잡한 동시 데이터 구조를 피함으로써 시스템 설계는 추론하기가 훨씬 간단해집니다. 이러한 단순성은 개별 주문 처리에 대한 보다 예측 가능하고 일관된 대기 시간으로 이어지며, 이는 저지연 환경에서 매우 중요합니다.
핵심 매칭 로직은 단일 스레드로 유지되지만 입출력 처리, 시장 데이터 배포 및 지속성과 같은 다른 시스템 구성 요소는 사용 가능한 CPU 코어를 최대한 활용하고 전체 시스템 처리량을 극대화하기 위해 다중 스레드화될 수 있고 또 그렇게 해야 합니다.

### 링 버퍼 최적화 및 그 역할(상세 설명)
초기 논의에서는 이 단일 스레드 아키텍처 내에서 링 버퍼(원형 버퍼 또는 SPSC 큐라고도 함)의 역할을 고려했습니다. Erik Rigtorp의 "처리량을 위한 링 버퍼 최적화" 기사에서는 캐시 라인 정렬(alignas(64)) 및 캐시된 인덱스를 통해 원자적 작업을 줄이는 것과 같은 주요 최적화를 강조하여 처리량을 크게 향상시킵니다(초당 550만 항목에서 1억 1,200만 항목으로). 이러한 원칙, 특히 캐시 라인 인식 및 잘못된 공유 최소화에 관한 원칙은 구문 및 제어 메커니즘이 다르지만(예: 작업의 경우 sync/atomic, 정렬의 경우 신중한 구조체 필드 순서 지정) Go에서도 동일하게 적용됩니다.

내 아이디어의 맥락에서 링 버퍼는 주로 일시적인 버퍼 역할을 합니다. 목적은 들어오는 주문의 작고 단기적인 급증을 흡수하여 최소한의 "버퍼" 시간을 제공하는 것입니다. 데이터가 메모리 할당 해제라는 의미에서 "삭제"되지 않는 것이 중요합니다. 대신 읽기 포인터가 진행되어 개념적으로 새 쓰기를 위한 슬롯을 해제합니다. 이 고정 크기의 빠른 경로 버퍼는 시스템이 지속적인 과부하를 인식하고 장기적인 저장 솔루션으로 작동하는 대신 외부 확장 메커니즘을 트리거하는 데 소중한 밀리초를 제공합니다.

### 선제적 확장 및 외부 큐잉: 복원력의 핵심
저의 세련된 전략은 내부 버퍼에 한계가 있다는 이해를 중심으로 합니다. 들어오는 이벤트의 비율이 처리 용량을 지속적으로 초과하면 아무리 많은 버퍼링도 결국 과부하 또는 데이터 손실을 막을 수 없습니다. 따라서 실시간 모니터링을 기반으로 한 선제적 확장으로 초점이 이동합니다.

세련된 전략은 다음을 강조합니다.

*   **일시적인 버퍼로서의 링 버퍼:** 설명된 바와 같이, 매칭 엔진 내의 인메모리 링 버퍼는 장기적인 오버플로 저장을 위한 것이 아닙니다. 확장 이벤트가 발생할 때까지 마이크로 버스트를 완화하고 매우 짧은 기간(밀리초에서 몇 초)의 버퍼를 제공하기 위한 것입니다.
*   **부하에 대한 지속적인 모니터링:** 시스템은 다음을 지속적으로 측정합니다.
    *   처리 속도: 현재 엔진의 초당 매칭 작업(Mops/s).
    *   큐 용량/깊이: 링 버퍼가 얼마나 찼는지.
    *   수신 이벤트 비율: 초당 얼마나 많은 새 이벤트가 도착하는지.
*   **선제적 확장 결정:** 수신 속도가 처리 속도를 지속적으로 앞지르기 시작하고 링 버퍼 깊이가 안전 임계값을 초과하면 메커니즘이 트리거되어 다음을 수행합니다.
    *   새롭고 잠재적으로 더 강력한(더 큰 CPU 인스턴스) 매칭 엔진을 가동합니다.
    *   (잠재적으로) 들어오는 트래픽을 새 인스턴스로 리디렉션하거나 사용 가능한 인스턴스 간에 분산합니다.
*   **인메모리 오버플로 저장소 없음:** 이는 오버플로를 위한 보조적이고 더 복잡한 데이터 구조를 관리할 필요가 없으므로 매칭 엔진의 내부 로직을 크게 단순화합니다.
이는 복원력과 성능을 위한 강력한 설계입니다.

### 외부 이벤트 시퀀싱 및 큐잉의 필수적인 역할(상세 설명)
이 외부 구성 요소는 고빈도 아키텍처의 신경계가 됩니다. 단순한 큐가 아닙니다. 모든 이벤트에 대한 단일 진실 공급원으로서 순서 지정, 내구성 및 확장 전략을 가능하게 합니다. 이벤트가 매칭 엔진에 직접 도달하는 대신 먼저 외부의 내구성이 뛰어나고 순서가 지정된 추가 전용 로그 또는 메시지 큐로 들어갑니다.

#### 이 외부 시스템의 주요 기능 및 이점:

*   **엄격한 이벤트 순서 지정:** 이것이 가장 중요합니다. 모든 이벤트(주문, 취소, 수정)는 시스템에서 수신된 정확한 시간순으로 기록되고 전달되어야 합니다. 이는 종종 매칭 엔진에 대한 단일 파티션, 전역적으로 정렬된 로그를 의미합니다(모든 상품에 대한 모든 이벤트가 서로 상대적으로 처리되도록 보장하기 때문).
*   **내구성:** 이벤트는 수신 즉시 디스크에 유지됩니다. 이는 매칭 엔진이 처리하기 전에도 시스템 장애 시 데이터 손실을 방지합니다.
*   **재현성/재생성:** 불변 로그이기 때문에 특정 시점부터 전체 이벤트 시퀀스를 "재생"할 수 있습니다. 이는 다음에 매우 유용합니다.
    *   재해 복구: 충돌 후 로그에서 매칭 엔진의 상태를 재구축합니다.
    *   감사 및 규정 준수: 언제 무슨 일이 있었는지 정확히 증명합니다.
    *   테스트: 과거 프로덕션 트래픽에 대해 새 매칭 엔진 버전을 실행합니다.
*   **디커플링:** 이벤트 생산자(예: 주문을 수신하는 API 게이트웨이)는 매칭 엔진과 분리됩니다. 로그에 쓰고 매칭 엔진은 로그에서 읽습니다. 이를 통해 생산자와 소비자를 독립적으로 확장할 수 있습니다.
*   **역압 처리:** 외부 큐는 상당한 버스트를 처리할 수 있습니다. 매칭 엔진 소비자가 뒤처지면 이벤트는 삭제되거나 생산자가 무기한 차단되는 대신 내구성 있는 큐에 쌓입니다.
#### 외부 시퀀싱 및 큐잉을 위한 훌륭한 접근 방식/기술:

*   **Apache Kafka / Confluent Kafka:** 처리량이 높고 내결함성이 있는 분산 스트리밍 플랫폼의 업계 표준입니다. 정렬되고 내구성 있으며 재생 가능한 로그(토픽)를 제공합니다. 수평적으로 쉽게 확장할 수 있습니다. 분산 소비를 위해 소비자 그룹을 지원합니다.
*   **Apache Pulsar:** Kafka와 유사하지만 더 나은 다중 테넌시와 더 유연한 메시징 모델(큐 및 스트림)로 종종 인용됩니다. 또한 정렬되고 내구성 있는 메시지를 제공합니다.
*   **NATS Streaming / LiftBridge (최신 NATS 버전의 NATS JetStream):** NATS(메시징에 매우 빠르고 가벼움) 위에 구축되었습니다. 영구적이고 순서가 지정된 메시지 로그를 제공합니다. 많은 사용 사례에서 Kafka보다 운영이 더 간단합니다.
*   **사용자 지정 영구 추가 전용 로그:** 궁극적인 제어와 가장 낮은 대기 시간을 위해 일부 HFT 회사는 종종 최적화된 SSD 또는 NVMe 장치에 직접 쓰는 자체 맞춤형 내구성 로그 솔루션을 구축합니다. 이것은 매우 복잡한 작업이며 일반적으로 가장 극단적인 요구 사항에만 해당됩니다.
#### 외부 큐잉을 사용한 확장 메커니즘:
외부 정렬 로그가 있으면 확장 전략이 훨씬 깔끔해집니다.

*   **모니터링 서비스:** 별도의 서비스가 선택한 메트릭(매칭 엔진 내부 큐 깊이, 외부 큐 백로그, 들어오는 메시지 속도)을 지속적으로 모니터링합니다.
*   **확장 트리거:** 임계값을 초과하면 이 서비스는 인프라(예: Kubernetes, 사용자 지정 오케스트레이션 스크립트)를 트리거하여 새 매칭 엔진 인스턴스를 프로비저닝합니다.
*   **상태 핸드오버(중요!):** 매칭 엔진은 상당한 상태(주문서)를 유지하므로 새 인스턴스는 외부 큐의 현재 끝에서 처리를 시작할 수 없습니다.
*   **스냅샷:** 현재 매칭 엔진은 내부 상태(예: 전체 주문서, 미결 주문, 마지막으로 처리된 이벤트 ID)의 일관된 스냅샷을 주기적으로 생성해야 합니다. 이 스냅샷은 내구성 있는 저장소에 기록됩니다.
*   **새 인스턴스 부트스트랩:** 새 매칭 엔진이 시작되면 먼저 최신 내구성 스냅샷을 로드합니다. 그런 다음 스냅샷이 생성된 후의 이벤트 ID부터 외부 로그에서 이벤트를 사용합니다. 이렇게 하면 상태를 정확하게 재구성하고 스냅샷 이후에 발생했지만 인수하기 전에 발생한 모든 이벤트를 처리할 수 있습니다.
*   **"컷오버" 또는 "드레인 앤 스위치":** 새 인스턴스가 라이브 스트림을 완전히 따라잡거나 충분히 가까워지면 이전 인스턴스를 정상적으로 드레이닝하거나(트래픽이 분할된 경우) 모든 새 들어오는 트래픽을 새롭고 더 강력한 인스턴스로 간단히 전환할 수 있습니다. 그러면 이전 인스턴스를 해제할 수 있습니다.
### CI/CD의 이점
이 외부 이벤트 큐잉 시스템은 매칭 엔진의 CI/CD를 크게 단순화합니다.

*   **블루/그린 배포:** 동일한 외부 이벤트 로그에서 사용하는 새 버전(그린)의 매칭 엔진을 가동합니다. 그린이 완전히 따라잡고 검증되면 모든 새 들어오는 트래픽을 그린으로 전환합니다. 그린이 실패하면 신속하게 블루로 다시 전환합니다.
*   **카나리아 릴리스:** 점진적으로 들어오는 이벤트의 작은 비율을 새 버전에 라우팅하여 전체 약정 없이 프로덕션에서 성능과 안정성을 테스트합니다.
*   **프로덕션 데이터를 사용한 자동화된 테스트:** 테스트 환경을 쉽게 가동하고 내구성 있는 로그에서 특정 과거 이벤트 시퀀스를 재생하여 현실적인 부하에서 새로운 기능이나 버그 수정을 철저히 테스트할 수 있습니다. 이는 복잡한 시스템 동작을 검증하는 데 매우 강력합니다.
*   **롤백:** 새 배포로 인해 문제가 발생하면 결함이 있는 버전을 중지하고 이전의 안정적인 버전을 시작하여 외부 이벤트 로그를 가리키도록 할 수 있습니다. 중단된 지점부터 또는 마지막 스냅샷부터 다시 재생하여 계속됩니다.
*   **단순화된 개발:** 개발자는 이벤트 로그의 하위 집합 또는 전체 재생에 대해 매칭 엔진의 로컬 인스턴스를 실행하여 일관성을 보장하고 디버깅을 더 쉽게 할 수 있습니다.
## 결론
최소한의 일시적인 인메모리 버퍼를 사용하고 강력한 입력과 선제적 확장을 위해 외부의 내구성 있고 엄격하게 정렬된 이벤트 시퀀싱/큐잉 시스템에 의존하는 나의 세련된 전략은 복원력 있고 고성능 금융 시스템을 구축하기 위한 황금 표준을 나타냅니다. 이 접근 방식을 통해 핵심 단일 스레드 매칭 엔진은 중요한 고속 로직에만 집중하고 영구 저장, 로드 밸런싱 및 동적 확장의 복잡성을 전문화된 외부 인프라에 오프로드할 수 있습니다. 이러한 관심사 분리는 데이터 무결성과 순서를 보장할 뿐만 아니라 강력한 CI/CD에 필요한 유연성과 변동하는 시장 수요에 원활하게 적응할 수 있는 기능을 제공하여 성능이 뛰어나고 운영적으로 건전한 시스템으로 이어집니다.
