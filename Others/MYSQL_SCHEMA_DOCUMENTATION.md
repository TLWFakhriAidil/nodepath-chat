# MySQL Database Schema Documentation

Generated on: 2025-07-30 12:38:25

## Tables (33 total)

### Table: `ai_campaign_progress`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| campaign_id | int(11) | NO | - | - | - |
| device_id | char(36) | NO | - | - | - |
| leads_sent | int(11) | YES | - | 0 | - |
| leads_failed | int(11) | YES | - | 0 | - |
| status | varchar(50) | YES | - | active | - |
| last_activity | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `ai_campaign_progress` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `campaign_id` int(11) NOT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `leads_sent` int(11) DEFAULT '0',
  `leads_failed` int(11) DEFAULT '0',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'active',
  `last_activity` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `broadcast_messages`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| user_id | char(36) | NO | - | - | - |
| device_id | char(36) | NO | - | - | - |
| campaign_id | int(11) | YES | - | - | - |
| sequence_id | char(36) | YES | - | - | - |
| recipient_phone | varchar(50) | NO | - | - | - |
| message_type | varchar(50) | NO | - | - | - |
| content | text | YES | - | - | - |
| media_url | text | YES | - | - | - |
| status | varchar(50) | YES | - | pending | - |
| error_message | text | YES | - | - | - |
| scheduled_at | timestamp | YES | - | - | - |
| sent_at | timestamp | YES | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| group_id | varchar(255) | YES | - | - | - |
| group_order | int(11) | YES | - | - | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| recipient_name | text | YES | - | - | - |
| sequence_stepid | char(36) | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `broadcast_messages` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `campaign_id` int(11) DEFAULT NULL,
  `sequence_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `recipient_phone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `media_url` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `error_message` text COLLATE utf8mb4_unicode_ci,
  `scheduled_at` timestamp NULL DEFAULT NULL,
  `sent_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `group_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `group_order` int(11) DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `recipient_name` text COLLATE utf8mb4_unicode_ci,
  `sequence_stepid` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `campaigns`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| user_id | char(36) | NO | - | - | - |
| campaign_date | date | NO | - | - | - |
| title | varchar(255) | NO | - | - | - |
| message | text | NO | - | - | - |
| image_url | text | YES | - | - | - |
| status | varchar(50) | YES | - | scheduled | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| niche | varchar(255) | YES | - | - | - |
| device_id | char(36) | YES | - | - | - |
| scheduled_time | varchar(255) | YES | - | - | - |
| min_delay_seconds | int(11) | YES | - | 10 | - |
| max_delay_seconds | int(11) | YES | - | 30 | - |
| target_status | varchar(100) | YES | - | customer | - |
| time_schedule | varchar(50) | YES | - | - | - |
| scheduled_at | timestamp | YES | - | - | - |
| ai | varchar(10) | YES | - | - | - |
| limit | int(11) | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `campaigns` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `campaign_date` date NOT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `image_url` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'scheduled',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `niche` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scheduled_time` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `min_delay_seconds` int(11) DEFAULT '10',
  `max_delay_seconds` int(11) DEFAULT '30',
  `target_status` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT 'customer',
  `time_schedule` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scheduled_at` timestamp NULL DEFAULT NULL,
  `ai` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `limit` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `device_load_balance`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| device_id | char(36) | NO | PRI | - | - |
| messages_hour | int(11) | YES | - | 0 | - |
| messages_today | int(11) | YES | - | 0 | - |
| last_reset_hour | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| last_reset_day | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| is_available | tinyint(1) | YES | - | 1 | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `device_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `device_load_balance` (
  `device_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `messages_hour` int(11) DEFAULT '0',
  `messages_today` int(11) DEFAULT '0',
  `last_reset_hour` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `last_reset_day` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `is_available` tinyint(1) DEFAULT '1',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `leads`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| device_id | char(36) | NO | - | - | - |
| user_id | char(36) | NO | - | - | - |
| name | varchar(255) | NO | - | - | - |
| phone | varchar(50) | NO | - | - | - |
| niche | varchar(255) | YES | - | - | - |
| journey | text | YES | - | - | - |
| status | varchar(50) | YES | - | new | - |
| last_interaction | timestamp | YES | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| target_status | varchar(100) | YES | - | customer | - |
| trigger | varchar(1000) | YES | - | - | - |
| provider | text | YES | - | - | - |
| platform | varchar(50) | YES | - | - | - |
| group | text | YES | - | - | - |
| community | text | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `leads` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `niche` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `journey` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'new',
  `last_interaction` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `target_status` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT 'customer',
  `trigger` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `provider` text COLLATE utf8mb4_unicode_ci,
  `platform` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `group` text COLLATE utf8mb4_unicode_ci,
  `community` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28711 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `leads_ai`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| user_id | char(36) | NO | - | - | - |
| device_id | char(36) | YES | - | - | - |
| name | varchar(255) | NO | - | - | - |
| phone | varchar(50) | NO | - | - | - |
| email | varchar(255) | YES | - | - | - |
| niche | varchar(255) | YES | - | - | - |
| source | varchar(255) | YES | - | ai_manual | - |
| status | varchar(50) | YES | - | pending | - |
| target_status | varchar(50) | YES | - | prospect | - |
| notes | text | YES | - | - | - |
| assigned_at | timestamp | YES | - | - | - |
| sent_at | timestamp | YES | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `leads_ai` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `niche` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `source` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'ai_manual',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `target_status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'prospect',
  `notes` text COLLATE utf8mb4_unicode_ci,
  `assigned_at` timestamp NULL DEFAULT NULL,
  `sent_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `message_analytics`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| user_id | char(36) | NO | - | - | - |
| device_id | char(36) | YES | - | - | - |
| message_id | varchar(255) | NO | - | - | - |
| jid | varchar(255) | NO | - | - | - |
| content | text | YES | - | - | - |
| is_from_me | tinyint(1) | YES | - | 0 | - |
| status | varchar(50) | NO | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `message_analytics` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `message_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `is_from_me` tinyint(1) DEFAULT '0',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `sequence_contacts`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| sequence_id | char(36) | NO | - | - | - |
| contact_phone | varchar(50) | NO | - | - | - |
| contact_name | varchar(255) | YES | - | - | - |
| current_step | int(11) | YES | - | 0 | - |
| status | varchar(50) | YES | - | active | - |
| completed_at | timestamp | YES | - | - | - |
| current_trigger | varchar(255) | YES | - | - | - |
| next_trigger_time | timestamp | YES | - | - | - |
| processing_device_id | char(36) | YES | - | - | - |
| last_error | text | YES | - | - | - |
| retry_count | int(11) | YES | - | 0 | - |
| assigned_device_id | char(36) | YES | - | - | - |
| processing_started_at | timestamp | YES | - | - | - |
| sequence_stepid | char(36) | NO | - | - | - |
| user_id | char(36) | YES | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `sequence_contacts` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sequence_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `contact_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `current_step` int(11) DEFAULT '0',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'active',
  `completed_at` timestamp NULL DEFAULT NULL,
  `current_trigger` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `next_trigger_time` timestamp NULL DEFAULT NULL,
  `processing_device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_error` text COLLATE utf8mb4_unicode_ci,
  `retry_count` int(11) DEFAULT '0',
  `assigned_device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `processing_started_at` timestamp NULL DEFAULT NULL,
  `sequence_stepid` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `sequence_logs`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| sequence_id | char(36) | NO | - | - | - |
| contact_id | char(36) | NO | - | - | - |
| step_id | char(36) | NO | - | - | - |
| day | int(11) | NO | - | - | - |
| status | varchar(50) | NO | - | - | - |
| message_id | varchar(255) | YES | - | - | - |
| error_message | text | YES | - | - | - |
| sent_at | timestamp | NO | - | CURRENT_TIMESTAMP | on update CURRENT_TIMESTAMP |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `sequence_logs` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sequence_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `contact_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `step_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `day` int(11) NOT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `error_message` text COLLATE utf8mb4_unicode_ci,
  `sent_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `sequence_steps`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| sequence_id | char(36) | NO | - | - | - |
| day_number | int(11) | NO | - | - | - |
| message_type | varchar(50) | NO | - | - | - |
| content | text | YES | - | - | - |
| media_url | text | YES | - | - | - |
| caption | text | YES | - | - | - |
| delay_days | int(11) | YES | - | 1 | - |
| time_schedule | varchar(50) | YES | - | - | - |
| trigger | text | YES | - | - | - |
| next_trigger | varchar(255) | YES | - | - | - |
| trigger_delay_hours | int(11) | YES | - | 24 | - |
| is_entry_point | tinyint(1) | YES | - | 0 | - |
| min_delay_seconds | int(11) | YES | - | 10 | - |
| max_delay_seconds | int(11) | YES | - | 30 | - |
| day | int(11) | YES | - | - | - |
| send_time | varchar(10) | YES | - | - | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `sequence_steps` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sequence_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `day_number` int(11) NOT NULL,
  `message_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `media_url` text COLLATE utf8mb4_unicode_ci,
  `caption` text COLLATE utf8mb4_unicode_ci,
  `delay_days` int(11) DEFAULT '1',
  `time_schedule` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `trigger` text COLLATE utf8mb4_unicode_ci,
  `next_trigger` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `trigger_delay_hours` int(11) DEFAULT '24',
  `is_entry_point` tinyint(1) DEFAULT '0',
  `min_delay_seconds` int(11) DEFAULT '10',
  `max_delay_seconds` int(11) DEFAULT '30',
  `day` int(11) DEFAULT NULL,
  `send_time` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `sequences`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| user_id | char(36) | NO | - | - | - |
| device_id | char(36) | YES | - | - | - |
| name | varchar(255) | NO | - | - | - |
| description | text | YES | - | - | - |
| niche | varchar(255) | YES | - | - | - |
| status | varchar(50) | YES | - | draft | - |
| auto_enroll | tinyint(1) | YES | - | 0 | - |
| skip_weekends | tinyint(1) | YES | - | 0 | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| total_days | int(11) | YES | - | 0 | - |
| is_active | tinyint(1) | YES | - | 1 | - |
| schedule_time | varchar(10) | YES | - | - | - |
| min_delay_seconds | int(11) | YES | - | 10 | - |
| max_delay_seconds | int(11) | YES | - | 30 | - |
| target_status | varchar(100) | YES | - | customer | - |
| time_schedule | varchar(50) | YES | - | - | - |
| total_contacts | int(11) | YES | - | 0 | - |
| active_contacts | int(11) | YES | - | 0 | - |
| completed_contacts | int(11) | YES | - | 0 | - |
| failed_contacts | int(11) | YES | - | 0 | - |
| progress_percentage | decimal(10,2) | YES | - | 0.00 | - |
| last_activity_at | timestamp | YES | - | - | - |
| estimated_completion_at | timestamp | YES | - | - | - |
| start_trigger | text | YES | - | - | - |
| end_trigger | text | YES | - | - | - |
| trigger | varchar(255) | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `sequences` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `device_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `niche` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'draft',
  `auto_enroll` tinyint(1) DEFAULT '0',
  `skip_weekends` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `total_days` int(11) DEFAULT '0',
  `is_active` tinyint(1) DEFAULT '1',
  `schedule_time` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `min_delay_seconds` int(11) DEFAULT '10',
  `max_delay_seconds` int(11) DEFAULT '30',
  `target_status` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT 'customer',
  `time_schedule` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `total_contacts` int(11) DEFAULT '0',
  `active_contacts` int(11) DEFAULT '0',
  `completed_contacts` int(11) DEFAULT '0',
  `failed_contacts` int(11) DEFAULT '0',
  `progress_percentage` decimal(10,2) DEFAULT '0.00',
  `last_activity_at` timestamp NULL DEFAULT NULL,
  `estimated_completion_at` timestamp NULL DEFAULT NULL,
  `start_trigger` text COLLATE utf8mb4_unicode_ci,
  `end_trigger` text COLLATE utf8mb4_unicode_ci,
  `trigger` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `team_members`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| username | varchar(255) | NO | - | - | - |
| password | varchar(255) | NO | - | - | - |
| created_by | char(36) | YES | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| is_active | tinyint(1) | YES | - | 1 | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `team_members` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_by` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `is_active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `team_sessions`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| team_member_id | char(36) | YES | - | - | - |
| token | varchar(255) | NO | - | - | - |
| expires_at | timestamp | NO | - | CURRENT_TIMESTAMP | on update CURRENT_TIMESTAMP |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `team_sessions` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `team_member_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `user_devices`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| user_id | char(36) | NO | - | - | - |
| device_name | varchar(255) | NO | - | - | - |
| phone | varchar(50) | YES | - | - | - |
| jid | varchar(255) | YES | - | - | - |
| status | varchar(50) | YES | - | offline | - |
| last_seen | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| min_delay_seconds | int(11) | YES | - | 5 | - |
| max_delay_seconds | int(11) | YES | - | 15 | - |
| wablas_instance | text | YES | - | - | - |
| whacenter_instance | text | YES | - | - | - |
| platform | varchar(50) | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `user_devices` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `device_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'offline',
  `last_seen` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `min_delay_seconds` int(11) DEFAULT '5',
  `max_delay_seconds` int(11) DEFAULT '15',
  `wablas_instance` text COLLATE utf8mb4_unicode_ci,
  `whacenter_instance` text COLLATE utf8mb4_unicode_ci,
  `platform` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `user_sessions`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| user_id | char(36) | NO | - | - | - |
| token | varchar(255) | NO | - | - | - |
| expires_at | timestamp | NO | - | CURRENT_TIMESTAMP | on update CURRENT_TIMESTAMP |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `user_sessions` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `users`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | char(36) | NO | PRI | - | - |
| email | varchar(255) | NO | - | - | - |
| full_name | varchar(255) | NO | - | - | - |
| password_hash | varchar(255) | NO | - | - | - |
| is_active | tinyint(1) | YES | - | 1 | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| updated_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| last_login | timestamp | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `users` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `full_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `last_login` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsapp_chats`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| device_id | varchar(255) | NO | - | - | - |
| chat_jid | varchar(255) | NO | - | - | - |
| last_message_time | timestamp | NO | - | CURRENT_TIMESTAMP | on update CURRENT_TIMESTAMP |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| chat_name | text | NO | - | - | - |
| updated_at | date | YES | - | - | - |
| is_group | tinyint(1) | YES | - | 0 | - |
| is_muted | tinyint(1) | YES | - | 0 | - |
| last_message_text | text | YES | - | - | - |
| unread_count | int(11) | YES | - | 0 | - |
| avatar_url | text | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsapp_chats` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `device_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chat_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `last_message_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `chat_name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `updated_at` date DEFAULT NULL,
  `is_group` tinyint(1) DEFAULT '0',
  `is_muted` tinyint(1) DEFAULT '0',
  `last_message_text` text COLLATE utf8mb4_unicode_ci,
  `unread_count` int(11) DEFAULT '0',
  `avatar_url` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsapp_messages`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| id | int(11) | NO | PRI | - | auto_increment |
| device_id | varchar(255) | NO | - | - | - |
| chat_jid | varchar(255) | NO | - | - | - |
| message_id | varchar(255) | NO | - | - | - |
| sender_jid | varchar(255) | YES | - | - | - |
| sender_name | varchar(255) | YES | - | - | - |
| message_text | text | YES | - | - | - |
| message_type | varchar(50) | YES | - | text | - |
| media_url | text | YES | - | - | - |
| is_sent | tinyint(1) | YES | - | 0 | - |
| is_read | tinyint(1) | YES | - | 0 | - |
| timestamp | bigint(20) | NO | - | - | - |
| created_at | timestamp | YES | - | CURRENT_TIMESTAMP | - |
| message_secrets | text | YES | - | - | - |

**Indexes:**
- PRIMARY on `id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsapp_messages` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `device_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chat_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sender_jid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sender_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `message_text` text COLLATE utf8mb4_unicode_ci,
  `message_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'text',
  `media_url` text COLLATE utf8mb4_unicode_ci,
  `is_sent` tinyint(1) DEFAULT '0',
  `is_read` tinyint(1) DEFAULT '0',
  `timestamp` bigint(20) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `message_secrets` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_app_state_mutation_macs`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| jid | varchar(100) | NO | PRI | - | - |
| name | varchar(255) | NO | PRI | - | - |
| version | bigint(20) | NO | PRI | - | - |
| index_mac | varbinary(255) | NO | PRI | - | - |
| value_mac | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `jid` (UNIQUE)
- PRIMARY on `name` (UNIQUE)
- PRIMARY on `version` (UNIQUE)
- PRIMARY on `index_mac` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_app_state_mutation_macs` (
  `jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `version` bigint(20) NOT NULL,
  `index_mac` varbinary(255) NOT NULL,
  `value_mac` longblob,
  PRIMARY KEY (`jid`,`name`,`version`,`index_mac`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_app_state_sync_keys`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| jid | varchar(255) | NO | PRI | - | - |
| key_id | varbinary(255) | NO | PRI | - | - |
| key_data | longblob | YES | - | - | - |
| timestamp | bigint(20) | YES | - | - | - |
| fingerprint | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `jid` (UNIQUE)
- PRIMARY on `key_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_app_state_sync_keys` (
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `key_id` varbinary(255) NOT NULL,
  `key_data` longblob,
  `timestamp` bigint(20) DEFAULT NULL,
  `fingerprint` longblob,
  PRIMARY KEY (`jid`,`key_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_app_state_version`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| jid | varchar(255) | NO | PRI | - | - |
| name | varchar(255) | NO | PRI | - | - |
| version | bigint(20) | YES | - | - | - |
| hash | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `jid` (UNIQUE)
- PRIMARY on `name` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_app_state_version` (
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `version` bigint(20) DEFAULT NULL,
  `hash` longblob,
  PRIMARY KEY (`jid`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_chat_settings`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(255) | NO | PRI | - | - |
| chat_jid | varchar(255) | NO | PRI | - | - |
| muted_until | bigint(20) | YES | - | - | - |
| pinned | tinyint(1) | YES | - | - | - |
| archived | tinyint(1) | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `chat_jid` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_chat_settings` (
  `our_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chat_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `muted_until` bigint(20) DEFAULT NULL,
  `pinned` tinyint(1) DEFAULT NULL,
  `archived` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`our_jid`,`chat_jid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_contacts`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(255) | NO | PRI | - | - |
| their_jid | varchar(255) | NO | PRI | - | - |
| first_name | text | YES | - | - | - |
| full_name | text | YES | - | - | - |
| push_name | text | YES | - | - | - |
| business_name | text | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `their_jid` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_contacts` (
  `our_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `their_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `first_name` text COLLATE utf8mb4_unicode_ci,
  `full_name` text COLLATE utf8mb4_unicode_ci,
  `push_name` text COLLATE utf8mb4_unicode_ci,
  `business_name` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`our_jid`,`their_jid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_device`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| jid | varchar(255) | NO | PRI | - | - |
| lid | varchar(255) | YES | - | - | - |
| registration_id | bigint(20) | NO | - | - | - |
| noise_key | longblob | NO | - | - | - |
| identity_key | longblob | NO | - | - | - |
| signed_pre_key | longblob | NO | - | - | - |
| signed_pre_key_id | int(11) | NO | - | - | - |
| signed_pre_key_sig | longblob | NO | - | - | - |
| adv_key | longblob | YES | - | - | - |
| adv_details | longblob | YES | - | - | - |
| adv_account_sig | longblob | YES | - | - | - |
| adv_account_sig_key | longblob | YES | - | - | - |
| adv_device_sig | longblob | YES | - | - | - |
| platform | varchar(50) | YES | - | - | - |
| business_name | text | YES | - | - | - |
| push_name | text | YES | - | - | - |
| facebook_uuid | text | YES | - | - | - |
| initialized | tinyint(1) | YES | - | 0 | - |
| account | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `jid` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_device` (
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `lid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `registration_id` bigint(20) NOT NULL,
  `noise_key` longblob NOT NULL,
  `identity_key` longblob NOT NULL,
  `signed_pre_key` longblob NOT NULL,
  `signed_pre_key_id` int(11) NOT NULL,
  `signed_pre_key_sig` longblob NOT NULL,
  `adv_key` longblob,
  `adv_details` longblob,
  `adv_account_sig` longblob,
  `adv_account_sig_key` longblob,
  `adv_device_sig` longblob,
  `platform` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `business_name` text COLLATE utf8mb4_unicode_ci,
  `push_name` text COLLATE utf8mb4_unicode_ci,
  `facebook_uuid` text COLLATE utf8mb4_unicode_ci,
  `initialized` tinyint(1) DEFAULT '0',
  `account` longblob,
  PRIMARY KEY (`jid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_event_buffer`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(255) | NO | PRI | - | - |
| ciphertext_hash | varbinary(255) | NO | PRI | - | - |
| plaintext | longblob | YES | - | - | - |
| server_timestamp | bigint(20) | NO | - | - | - |
| insert_timestamp | bigint(20) | NO | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `ciphertext_hash` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_event_buffer` (
  `our_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ciphertext_hash` varbinary(255) NOT NULL,
  `plaintext` longblob,
  `server_timestamp` bigint(20) NOT NULL,
  `insert_timestamp` bigint(20) NOT NULL,
  PRIMARY KEY (`our_jid`,`ciphertext_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_identity_keys`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(255) | NO | PRI | - | - |
| their_id | varchar(255) | NO | PRI | - | - |
| identity | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `their_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_identity_keys` (
  `our_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `their_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `identity` longblob,
  PRIMARY KEY (`our_jid`,`their_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_lid_map`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| lid | varchar(255) | NO | PRI | - | - |
| pn | text | NO | - | - | - |

**Indexes:**
- PRIMARY on `lid` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_lid_map` (
  `lid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pn` text COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`lid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_message_secrets`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(100) | NO | PRI | - | - |
| chat_jid | varchar(100) | NO | PRI | - | - |
| sender_jid | varchar(100) | NO | PRI | - | - |
| message_id | varchar(100) | NO | PRI | - | - |
| secret | longblob | YES | - | - | - |
| key | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `chat_jid` (UNIQUE)
- PRIMARY on `sender_jid` (UNIQUE)
- PRIMARY on `message_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_message_secrets` (
  `our_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chat_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sender_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `secret` longblob,
  `key` longblob,
  PRIMARY KEY (`our_jid`,`chat_jid`,`sender_jid`,`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_pre_keys`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| jid | varchar(255) | NO | PRI | - | - |
| key_id | int(11) | NO | PRI | - | - |
| key | longblob | YES | - | - | - |
| uploaded | tinyint(1) | YES | - | - | - |

**Indexes:**
- PRIMARY on `jid` (UNIQUE)
- PRIMARY on `key_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_pre_keys` (
  `jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `key_id` int(11) NOT NULL,
  `key` longblob,
  `uploaded` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`jid`,`key_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_privacy_tokens`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(100) | NO | PRI | - | - |
| their_jid | varchar(100) | NO | PRI | - | - |
| token | longblob | YES | - | - | - |
| timestamp | bigint(20) | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `their_jid` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_privacy_tokens` (
  `our_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `their_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `token` longblob,
  `timestamp` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`our_jid`,`their_jid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_sender_keys`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(100) | NO | PRI | - | - |
| chat_id | varchar(255) | NO | PRI | - | - |
| sender_id | varchar(255) | NO | PRI | - | - |
| sender_key | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `chat_id` (UNIQUE)
- PRIMARY on `sender_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_sender_keys` (
  `our_jid` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chat_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sender_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sender_key` longblob,
  PRIMARY KEY (`our_jid`,`chat_id`,`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_sessions`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| our_jid | varchar(255) | NO | PRI | - | - |
| their_id | varchar(255) | NO | PRI | - | - |
| session | longblob | YES | - | - | - |

**Indexes:**
- PRIMARY on `our_jid` (UNIQUE)
- PRIMARY on `their_id` (UNIQUE)

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_sessions` (
  `our_jid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `their_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `session` longblob,
  PRIMARY KEY (`our_jid`,`their_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

### Table: `whatsmeow_version`

| Column | Type | Null | Key | Default | Extra |
|--------|------|------|-----|---------|-------|
| version | int(11) | YES | - | - | - |
| compat | int(11) | YES | - | - | - |

<details>
<summary>CREATE Statement</summary>

```sql
CREATE TABLE `whatsmeow_version` (
  `version` int(11) DEFAULT NULL,
  `compat` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```
</details>

---

