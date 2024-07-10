CREATE DATABASE IF NOT EXISTS social_server;

CREATE TABLE social_server.tb_users (
    user_id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    password VARCHAR(100) NOT NULL,
    username VARCHAR(50) NOT NULL COLLATE utf8_general_ci UNIQUE,
    nickname VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE,
    avatar VARCHAR(100) DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE social_server.tb_user_contacts (
    user_id BIGINT UNSIGNED,
    contact_id BIGINT UNSIGNED,
    is_mutual_contact BOOLEAN DEFAULT FALSE,        -- 是否为双向联系人
    remark_name VARCHAR(50) DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, contact_id),
    FOREIGN KEY (user_id) REFERENCES tb_users(user_id)
);

CREATE TABLE social_server.tb_groups (
    group_id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    group_name VARCHAR(100) NOT NULL,
    owner_id BIGINT UNSIGNED NOT NULL,
    avatar VARCHAR(100) DEFAULT '',
    mem_count INT UNSIGNED DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES tb_users(user_id)
);

CREATE TABLE social_server.tb_group_members (
    group_id BIGINT UNSIGNED,
    user_id BIGINT UNSIGNED,
    role INT UNSIGNED DEFAULT 0,        -- 0 普通成员, 1 群主, 2 管理员
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES tb_groups(group_id),
    FOREIGN KEY (user_id) REFERENCES tb_users(user_id)
);

CREATE TABLE social_server.tb_user_inbox (
    user_id BIGINT UNSIGNED NOT NULL,
	seq_id BIGINT UNSIGNED NOT NULL,

    sender_id BIGINT UNSIGNED NOT NULL,
    receiver_id BIGINT UNSIGNED,
    group_id BIGINT UNSIGNED,

    conv_msg_id BIGINT UNSIGNED NOT NULL,
	rand_msg_id BIGINT UNSIGNED DEFAULT 0,
    message_type INT NOT NULL,
    content TEXT NOT NULL,
    read_msg_id BIGINT UNSIGNED DEFAULT 0,

    is_read BOOLEAN DEFAULT FALSE,
    status INT DEFAULT 0,       -- 好友/加群申请：0 未处理, 1 同意, 2 拒绝, 3 忽略

    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY (user_id, seq_id)
);

CREATE TABLE social_server.tb_seq_id_user (
	user_id BIGINT UNSIGNED,
	seq_id BIGINT UNSIGNED,
	PRIMARY KEY (user_id, seq_id)
);

CREATE TABLE social_server.tb_seq_id_chat (
	user1_id BIGINT UNSIGNED,
	user2_id BIGINT UNSIGNED,
	seq_id BIGINT UNSIGNED,
	PRIMARY KEY (user1_id, user2_id, seq_id)
);

CREATE TABLE social_server.tb_seq_id_group (
	group_id BIGINT UNSIGNED,
	seq_id BIGINT UNSIGNED,
	PRIMARY KEY (group_id, seq_id)
);
