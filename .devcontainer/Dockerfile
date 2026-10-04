# miseの公式Debianイメージ．git，curl，CA証明書とmiseを含み，
# PATHに/mise/shims，MISE_DATA_DIRに/miseが設定済み．
FROM jdxcode/mise:2026.10.0-debian

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
	openssh-client \
	procps \
	jq \
	libatomic1 \
	postgresql-client \
	&& rm -rf /var/lib/apt/lists/*

# 講座の道具は，mise.tomlとmise.lockのとおりの版とチェックサムでイメージへ入れる．
WORKDIR /tmp/mise-bootstrap
COPY mise.toml mise.lock ./
# Pythonで書かれた道具(semgrep，awscli)の依存関係の固定
COPY .mise/locks ./.mise/locks
RUN mise trust \
	&& mise ls --current \
	&& mise install --locked

# rtkのhookを有効化する．
RUN mkdir -p /root/.claude \
	&& rtk init -g --hook-only

# 起動後は，リポジトリのmise.tomlをグローバル設定として読む．
# これにより，/workspacesの外のディレクトリでも同じツールを使える．
# グローバル設定のルートも/workspacesにして，タスクをリポジトリ直下で実行させる
# (未指定だとホームディレクトリで実行される)．
ENV MISE_GLOBAL_CONFIG_FILE=/workspaces/mise.toml \
	MISE_GLOBAL_CONFIG_ROOT=/workspaces \
	MISE_TRUSTED_CONFIG_PATHS=/workspaces

WORKDIR /workspaces
