pipeline {
    agent any

    environment {
        GO_VERSION = '1.21'
        DEPLOY_PATH = '/var/wxpush'
        BINARY_NAME = 'wxpush'
        TENCENT_NODE_DEPLOY_USER = 'ubuntu'
        TENCENT_NODE_SSH_KEY_CREDENTIAL = 'TencentNodeSSHKey'
        VERSION = "${BUILD_NUMBER}"
        MAX_BACKUPS = 10
    }

    stages {
        stage('Checkout') {
            steps {
                echo '拉取代码...'
                checkout scm
                script {
                    echo "当前构建版本: ${VERSION}"
                    echo "当前构建时间: ${new Date().format('yyyy-MM-dd HH:mm:ss')}"
                }
            }
        }

        stage('Build') {
            steps {
                echo '构建项目...'
                withEnv(["PATH+GO=/usr/local/go/bin"]) {
                    sh 'go build -ldflags "-X main.Version=${VERSION}" -o ${BINARY_NAME} main.go'
                    sh 'ls -lh ${BINARY_NAME}'
                }
            }
        }

        stage('Copy to local server') {
            steps {
                script {
                    deployToLocal()
                }
            }
        }

        stage('Copy to remote server') {
            steps {
                script {
                    withCredentials([
                        string(credentialsId: 'TencentNodeIP', variable: 'DEPLOY_HOST'),
                        sshUserPrivateKey(credentialsId: TENCENT_NODE_SSH_KEY_CREDENTIAL, keyFileVariable: 'SSH_KEY')
                    ]) {
                        deployToRemote(DEPLOY_HOST, SSH_KEY)
                    }
                }
            }
        }

        stage('Cleanup') {
            steps {
                echo '清理临时文件...'
                cleanWs()
            }
        }
    }

    post {
        success {
            echo '构建和部署成功！'
            echo "可执行文件: ${BINARY_NAME}"
            echo "版本号: ${VERSION}"
            echo "部署路径: ${DEPLOY_PATH}"
            echo "最大备份数: ${MAX_BACKUPS}"
        }

        failure {
            echo '构建或部署失败！'
            echo "版本号: ${VERSION}"
        }

        always {
            echo '流水线执行完成'
        }
    }
}

def deployToLocal() {
    sh """
        # 复制新二进制文件到待部署文件夹
        echo "复制新二进制文件 ${BINARY_NAME} 到待部署文件夹 ${DEPLOY_PATH}_new..."
        sudo rm -rf ${DEPLOY_PATH}_new && sudo mkdir -p ${DEPLOY_PATH}_new
        sudo cp ${BINARY_NAME} ${DEPLOY_PATH}_new/
        
        ${deploy('local')}
    """
}

def deployToRemote(host, sshKey) {
    sh """
        set -e
        REMOTE="${TENCENT_NODE_DEPLOY_USER}@${host}"
        echo "连接远程服务器: \$REMOTE"
        echo "部署版本: ${VERSION}"
        
        # 在远程服务器删除旧目录并创建新目录
        ssh -i "${sshKey}" -o StrictHostKeyChecking=no "\$REMOTE" "sudo rm -rf ${DEPLOY_PATH}_new && sudo mkdir -p ${DEPLOY_PATH}_new"
        
        # 同步二进制文件到远程服务器
        echo "同步二进制文件..."
        rsync -avz --delete --rsync-path="sudo rsync" -e "ssh -i ${sshKey} -o StrictHostKeyChecking=no" ${BINARY_NAME} "\$REMOTE:${DEPLOY_PATH}_new/"
        
        # 在远程服务器执行部署脚本
        ssh -i "${sshKey}" -o StrictHostKeyChecking=no "\$REMOTE" "${deploy('remote')}"
    """
}

def deploy(type) {
    """
        # 定义备份目录名称（带版本号后缀）
        BACKUP_DIR="${DEPLOY_PATH}_backup_v${VERSION}"
        
        # 备份旧版本
        ${backupOldVersion(type)}
        
        # 使用 rsync 同步新版本到部署路径（自动处理新增、修改和删除的文件）
        sudo rsync -avz --delete ${DEPLOY_PATH}_new/ ${DEPLOY_PATH}/
        
        # 删除临时目录
        sudo rm -rf ${DEPLOY_PATH}_new
        
        # 设置执行权限
        sudo chmod +x ${DEPLOY_PATH}/${BINARY_NAME}
        
        # 清理旧备份
        ${cleanupOldBackups(type)}
        
        # 验证部署
        echo ""
        echo "已部署文件："
        ls -lh ${DEPLOY_PATH}/
    """
}

def backupOldVersion(type) {
    def backupScript = """
        if [ -d '${DEPLOY_PATH}' ]; then
            sudo cp -r ${DEPLOY_PATH} '${DEPLOY_PATH}_backup_v${VERSION}'
            echo '已备份旧版本'
        else
            echo '首次部署，无需备份'
        fi
    """
    
    if (type == 'local') {
        return """
            echo "备份旧版本（如果存在）..."
            ${backupScript}
        """
    } else {
        return """
            echo "备份旧版本到 ${DEPLOY_PATH}_backup_v${VERSION}..."
            ${backupScript}
        """
    }
}

def cleanupOldBackups(type) {
    def cleanupScript = """
        BACKUP_COUNT=\$(ls -d ${DEPLOY_PATH}_backup_v* 2>/dev/null | wc -l | tr -d ' ')
        echo '当前备份数量: '\$BACKUP_COUNT
        
        if [ "\${BACKUP_COUNT:-0}" -gt ${MAX_BACKUPS} ]; then
            DELETE_COUNT=\$((BACKUP_COUNT - MAX_BACKUPS))
            echo '需要删除 '\$DELETE_COUNT' 个旧备份'
            ls -dt ${DEPLOY_PATH}_backup_v* | tail -n \$DELETE_COUNT | while read dir; do
                echo '删除旧备份: '\$dir
                sudo rm -rf '\$dir'
            done
        else
            echo '备份数量在限制范围内，无需清理'
        fi
        
        echo ''
        echo '当前备份列表：'
        ls -lhdt ${DEPLOY_PATH}_backup_v* 2>/dev/null || echo '暂无备份'
    """
    
    if (type == 'local') {
        return """
            echo "清理旧备份文件..."
            ${cleanupScript}
        """
    } else {
        return """
            echo "清理旧备份文件..."
            ${cleanupScript}
        """
    }
}
