use std::sync::Arc;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::{self, KeyAlgorithm, KeyUsage, KeyStatus, KeyMetadata};

/// Simulated HSM key manager for development/testing
/// In production, this would integrate with actual HSM hardware via PKCS#11
pub struct SoftHsmKeyManager {
    /// Internal key storage (simulated)
    keys: Arc<dashmap::DashMap<String, KeyMetadata>>,
    /// Key storage for public keys
    public_keys: Arc<dashmap::DashMap<String, Vec<u8>>>,
}

impl SoftHsmKeyManager {
    /// Creates a new SoftHSM key manager
    pub fn new() -> Self {
        Self {
            keys: Arc::new(dashmap::DashMap::new()),
            public_keys: Arc::new(dashmap::DashMap::new()),
        }
    }

    /// Generates a simulated key ID
    fn generate_key_id(&self, label: &str) -> String {
        format!("soft_hsm_{}_{}", label, Uuid::new_v4().to_simple())
    }
}

#[async_trait]
impl domain::KeyManager for SoftHsmKeyManager {
    async fn generate_key(
        &self,
        label: &str,
        algorithm: KeyAlgorithm,
    ) -> domain::DomainResult<KeyMetadata> {
        // Simulate key generation delay
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;

        let key_id = self.generate_key_id(label);
        let now = Utc::now();

        // Generate mock key material based on algorithm
        let (private_key, public_key) = match algorithm {
            KeyAlgorithm::EcdsaSecp256k1 => {
                // Mock secp256k1 key generation
                let private_key = rand::random::<[u8; 32]();
                let public_key = secp256k1::PublicKey::from_secret_key(
                    &secp256k1::Secp256k1::new(),
                    &secp256k1::PrivateKey::from_slice(&private_key).unwrap(),
                );
                (private_key.to_vec(), public_key.serialize().to_vec())
            }
            KeyAlgorithm::EcdsaP256 => {
                // Mock P-256 key generation
                let private_key = rand::random::<[u8; 32]>();
                let public_key = vec![4]; // Uncompressed point prefix
                public_key.extend_from_slice(&private_key[0..32]);
                public_key.extend_from_slice(&private_key[16..32]);
                (private_key.to_vec(), public_key)
            }
            KeyAlgorithm::EdDSA => {
                // Mock Ed25519 key generation
                let private_key = rand::random::<u64>();
                let public_key = private_key.to_le_bytes().to_vec();
                (private_key.to_le_bytes().to_vec(), public_key)
            }
            _ => {
                // Generic fallback
                let private_key = rand::random::<[u8; 32]>();
                (private_key.to_vec(), private_key[0..32].to_vec())
            }
        };

        // Derive usage from algorithm
        let usage = match algorithm {
            KeyAlgorithm::EcdsaSecp256k1 | KeyAlgorithm::EcdsaP256 | KeyAlgorithm::EdDSA => KeyUsage::Signing,
            _ => KeyUsage::Encryption,
        };

        let metadata = KeyMetadata {
            id: Uuid::new_v4(),
            label: label.to_string(),
            algorithm,
            usage,
            hsm_slot_id: "soft_hsm_slot_0".to_string(),
            hsm_key_handle: key_id.clone(),
            public_key: public_key.clone(),
            status: KeyStatus::Active,
            wallet_id: None,
            version: 1,
            created_at: now,
            activated_at: Some(now),
            expires_at: None,
            last_used_at: None,
            rotation_policy_id: None,
            usage_count: 0,
        };

        // Store key (in real HSM, private key never leaves the module)
        self.keys.insert(key_id.clone(), metadata.clone());
        self.public_keys.insert(key_id, public_key);

        Ok(metadata)
    }

    async fn sign(
        &self,
        key_handle: &str,
        data: &[u8],
    ) -> domain::DomainResult<Vec<u8>> {
        // Simulate signing delay
        tokio::time::sleep(tokio::time::Duration::from_millis(50)).await;

        // Check if key exists
        if !self.keys.contains_key(key_handle) {
            return Err(domain::DomainError::KeyNotFound(key_handle.to_string()));
        }

        // Get key metadata
        let mut metadata = self.keys.get(key_handle)
            .ok_or_else(|| domain::DomainError::KeyNotFound(key_handle.to_string()))?
            .value()
            .clone();

        // Check key status
        if metadata.status != KeyStatus::Active {
            return Err(domain::DomainError::KeyNotActive);
        }

        // Generate mock signature based on algorithm
        let signature = match metadata.algorithm {
            KeyAlgorithm::EcdsaSecp256k1 | KeyAlgorithm::EcdsaP256 => {
                // Mock ECDSA signature (DER format)
                let r = rand::random::<[u8; 32]>();
                let s = rand::random::<[u8; 32]>();
                
                // Construct DER-encoded signature
                let mut der = vec![0x30]; // SEQUENCE tag
                der.push(0x44); // Length (68 bytes)
                der.push(0x02); // INTEGER tag for r
                der.push(0x20); // Length (32 bytes)
                der.extend_from_slice(&r);
                der.push(0x02); // INTEGER tag for s
                der.push(0x20); // Length (32 bytes)
                der.extend_from_slice(&s);
                der
            }
            KeyAlgorithm::EdDSA => {
                // Mock Ed25519 signature
                rand::random::<[u8; 64]>().to_vec()
            }
            _ => {
                // Generic signature
                let mut sig = vec![0u8; 64];
                rand::thread_rng().fill_bytes(&mut sig);
                sig
            }
        };

        // Update usage metadata
        metadata.usage_count += 1;
        metadata.last_used_at = Some(Utc::now());
        self.keys.insert(key_handle.to_string(), metadata);

        Ok(signature)
    }

    async fn verify(
        &self,
        public_key: &[u8],
        data: &[u8],
        signature: &[u8],
    ) -> domain::DomainResult<bool> {
        // In a real implementation, this would verify using the actual cryptographic library
        // For simulation, we just check that all inputs are non-empty
        if public_key.is_empty() || data.is_empty() || signature.is_empty() {
            return Ok(false);
        }

        // Basic structural validation
        match signature.len() {
            64 | 70..=100 => Ok(true), // Typical signature lengths
            _ => Ok(false),
        }
    }

    async fn get_public_key(&self, key_handle: &str) -> domain::DomainResult<Vec<u8>> {
        self.public_keys.get(key_handle)
            .map(|k| k.value().clone())
            .ok_or_else(|| domain::DomainError::KeyNotFound(key_handle.to_string()))
    }

    async fn rotate_key(&self, key_handle: &str) -> domain::DomainResult<KeyMetadata> {
        // Mark old key as rotated
        if let Some(mut old_metadata) = self.keys.get_mut(key_handle) {
            old_metadata.status = KeyStatus::Rotated;
        }

        // Generate new key with same label
        let old_metadata = self.keys.get(key_handle)
            .ok_or_else(|| domain::DomainError::KeyNotFound(key_handle.to_string()))?
            .value()
            .clone();

        // Generate new key
        self.generate_key(&old_metadata.label, old_metadata.algorithm).await
    }

    async fn destroy_key(&self, key_handle: &str) -> domain::DomainResult<()> {
        // Mark key as destroyed
        if let Some(mut metadata) = self.keys.get_mut(key_handle) {
            metadata.status = KeyStatus::Destroyed;
        }

        // Remove from public keys
        self.public_keys.remove(key_handle);

        Ok(())
    }
}

/// Kafka message bus implementation
pub struct KafkaMessageBus {
    /// Kafka producer (simulated)
    producers: Arc<dashmap::DashMap<String, Vec<u8>>>,
    /// Topics
    topics: Vec<String>,
}

impl KafkaMessageBus {
    /// Creates a new Kafka message bus
    pub fn new() -> Self {
        Self {
            producers: Arc::new(dashmap::DashMap::new()),
            topics: vec![
                "wallet.events".to_string(),
                "signing.events".to_string(),
                "emergency.events".to_string(),
                "compliance.reports".to_string(),
            ],
        }
    }
}

#[async_trait]
impl crate::application::MessageBus for KafkaMessageBus {
    async fn publish(&self, topic: &str, key: &str, value: &[u8]) -> domain::DomainResult<()> {
        // Simulate Kafka publish
        self.producers.insert(
            format!("{}_{}", topic, key),
            value.to_vec(),
        );
        Ok(())
    }

    async fn publish_wallet_event(
        &self,
        event: &crate::application::WalletEvent,
    ) -> domain::DomainResult<()> {
        let payload = serde_json::to_vec(event)?;
        let key = match event {
            crate::application::WalletEvent::Created(id, _, _, _) => id.to_string(),
            crate::application::WalletEvent::Frozen(id, _) => id.to_string(),
            crate::application::WalletEvent::Unfrozen(id) => id.to_string(),
            crate::application::WalletEvent::Archived(id) => id.to_string(),
            crate::application::WalletEvent::Recovered(id) => id.to_string(),
            crate::application::WalletEvent::LimitsUpdated(id, _, _) => id.to_string(),
        };
        self.publish("wallet.events", &key, &payload).await
    }

    async fn publish_signing_event(
        &self,
        event: &crate::application::SigningEvent,
    ) -> domain::DomainResult<()> {
        let payload = serde_json::to_vec(event)?;
        let key = match event {
            crate::application::SigningEvent::RequestCreated(_, request_id, _) => request_id.to_string(),
            crate::application::SigningEvent::RequestApproved(_, request_id) => request_id.to_string(),
            crate::application::SigningEvent::RequestRejected(_, request_id, _) => request_id.to_string(),
            crate::application::SigningEvent::RequestCompleted(request_id) => request_id.to_string(),
            crate::application::SigningEvent::RequestTimedOut(request_id) => request_id.to_string(),
        };
        self.publish("signing.events", &key, &payload).await
    }

    async fn publish_emergency_event(
        &self,
        event: &crate::application::EmergencyEvent,
    ) -> domain::DomainResult<()> {
        let payload = serde_json::to_vec(event)?;
        let key = match event {
            crate::application::EmergencyEvent::FreezeInitiated(id, _) => id.to_string(),
            crate::application::EmergencyEvent::FreezeCompleted(id, _) => id.to_string(),
            crate::application::EmergencyEvent::UnfreezeInitiated(id, _) => id.to_string(),
            crate::application::EmergencyEvent::UnfreezeCompleted(id, _) => id.to_string(),
            crate::application::EmergencyEvent::AccessRequested(id, _) => id.to_string(),
            crate::application::EmergencyEvent::AccessApproved(_, admin_id) => admin_id.to_string(),
            crate::application::EmergencyEvent::AccessDenied(id, _) => id.to_string(),
        };
        self.publish("emergency.events", &key, &payload).await
    }
}

/// In-memory audit log implementation for development
pub struct InMemoryAuditLog {
    /// Audit events storage
    events: Arc<dashmap::DashMap<String, Vec<AuditEvent>>>,
}

impl InMemoryAuditLog {
    /// Creates a new in-memory audit log
    pub fn new() -> Self {
        Self {
            events: Arc::new(dashmap::DashMap::new()),
        }
    }
}

#[async_trait]
impl AuditLog for InMemoryAuditLog {
    async fn log(&self, event: &AuditEvent) -> domain::DomainResult<()> {
        let timestamp = event.timestamp.to_rfc3339();
        self.events
            .entry(timestamp)
            .or_insert_with(Vec::new)
            .push(event.clone());
        Ok(())
    }

    async fn log_batch(&self, events: &[AuditEvent]) -> domain::DomainResult<()> {
        for event in events {
            self.log(event).await?;
        }
        Ok(())
    }

    async fn query(
        &self,
        query: crate::domain::AuditQuery,
    ) -> domain::DomainResult<crate::domain::AuditQueryResult> {
        let mut all_events: Vec<AuditEvent> = self.events
            .iter()
            .flat_map(|entry| entry.value().clone())
            .collect();

        // Apply filters
        all_events.retain(|event| {
            // Category filter
            if !query.categories.is_empty() && !query.categories.contains(&event.category) {
                return false;
            }
            
            // Event type filter
            if !query.event_types.is_empty() && !query.event_types.contains(&event.event_type) {
                return false;
            }
            
            // Severity filter
            if !query.severities.is_empty() && !query.severities.contains(&event.severity) {
                return false;
            }
            
            // Actor filter
            if let Some(actor_id) = &query.actor_id {
                if event.actor.id != *actor_id {
                    return false;
                }
            }
            
            // Resource filter
            if let Some(resource_id) = &query.resource_id {
                if event.resource_id != Some(*resource_id) {
                    return false;
                }
            }
            
            // Time range filter
            if let Some(start) = query.start_time {
                if event.timestamp < start {
                    return false;
                }
            }
            if let Some(end) = query.end_time {
                if event.timestamp > end {
                    return false;
                }
            }
            
            // Success filter
            if let Some(success) = query.success {
                if event.success != success {
                    return false;
                }
            }

            true
        });

        // Sort events
        all_events.sort_by(|a, b| match query.sort_direction {
            crate::domain::SortDirection::Desc => b.timestamp.cmp(&a.timestamp),
            crate::domain::SortDirection::Asc => a.timestamp.cmp(&b.timestamp),
        });

        // Pagination
        let total_count = all_events.len() as u64;
        let page_size = query.page_size.max(1).min(1000) as usize;
        let start = ((query.page.max(1) - 1) as usize).saturating_mul(page_size);
        let events: Vec<AuditEvent> = all_events.into_iter().skip(start).take(page_size).collect();

        let total_pages = (total_count as f64 / page_size as f64).ceil() as u32;

        Ok(crate::domain::AuditQueryResult {
            events,
            total_count,
            page: query.page,
            page_size: page_size as u32,
            total_pages,
            has_next: query.page < total_pages,
            has_previous: query.page > 1,
        })
    }
}
