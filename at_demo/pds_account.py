import os
import sys
import json
import base64
import secrets
import requests
from typing import Optional, List, Dict
from dataclasses import dataclass
from dotenv import load_dotenv

@dataclass
class PDSConfig:
    hostname: str
    port: int
    admin_password: str

class PDSError(Exception):
    pass

class PDSAccount:
    # def __init__(self, env_file: str = "/home/ubuntu/workspace/avatar-ai/pds/service/pds.env"):
    #def __init__(self, env_file: str = "/home/ubuntu/workspace/avatar-ai/engine/at_demo/pds.env"):
    def __init__(self, env_file: str = "/home/zhongshangwu/workspace/avatar-ai/engine/at_demo/pds.env"):
        load_dotenv(env_file)

        hostname = os.getenv("PDS_HOSTNAME", "")
        port = int(os.getenv("PDS_PORT", ""))
        passwd = os.getenv("PDS_ADMIN_PASSWORD", "")
        hostname = "127.0.0.1"
        port = 2583
        # passwd = "admin-pass"
        passwd = "26b14a9defe96407d490ed54b8857f55"

        self.config = PDSConfig(
            hostname=hostname,
            port=port,
            admin_password=passwd
        )

        if self.config.port != 443:
            self.base_url = f"http://{self.config.hostname}:{self.config.port}/xrpc"
        else:
            self.base_url = f"https://{self.config.hostname}/xrpc"
        self.headers = {
            "Content-Type": "application/json"
        }
        self.auth = ("admin", self.config.admin_password)

    def _make_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
        url = f"{self.base_url}/{endpoint}"
        response = requests.request(
            method=method,
            url=url,
            headers=self.headers,
            auth=self.auth,
            json=data
        )
        response.raise_for_status()
        return response.json()

    def list_accounts(self) -> List[Dict]:
        """列出所有账户"""
        repos = self._make_request("GET", "com.atproto.sync.listRepos")
        accounts = []
        for repo in repos.get("repos", []):
            did = repo.get("did")
            if did:
                account_info = self._make_request(
                    "GET",
                    f"com.atproto.admin.getAccountInfo?did={did}"
                )
                accounts.append(account_info)
        return accounts

    def create_account(self, email: str, handle: str) -> Dict:
        """创建新账户"""
        # 生成随机密码
        password = base64.b64encode(secrets.token_bytes(30)).decode('utf-8')
        password = ''.join(c for c in password if c.isalnum())[:24]

        # 创建邀请码
        invite_code = self._make_request(
            "POST",
            "com.atproto.server.createInviteCode",
            {"useCount": 1}
        ).get("code")

        # 创建账户
        result = self._make_request(
            "POST",
            "com.atproto.server.createAccount",
            {
                "email": email,
                "handle": handle,
                "password": password,
                "inviteCode": invite_code
            }
        )

        if not result.get("did", "").startswith("did:"):
            raise PDSError(f"创建账户失败: {result.get('message', '未知错误')}")

        result["password"] = password  # 添加密码到返回结果
        return result

    def delete_account(self, did: str) -> None:
        """删除账户"""
        if not did.startswith("did:"):
            raise ValueError("DID 必须以 'did:' 开头")

        self._make_request(
            "POST",
            "com.atproto.admin.deleteAccount",
            {"did": did}
        )

    def takedown_account(self, did: str) -> None:
        """封禁账户"""
        if not did.startswith("did:"):
            raise ValueError("DID 必须以 'did:' 开头")

        payload = {
            "subject": {
                "$type": "com.atproto.admin.defs#repoRef",
                "did": did
            },
            "takedown": {
                "applied": True,
                "ref": str(int(time.time()))
            }
        }

        self._make_request(
            "POST",
            "com.atproto.admin.updateSubjectStatus",
            payload
        )

    def untakedown_account(self, did: str) -> None:
        """解除账户封禁"""
        if not did.startswith("did:"):
            raise ValueError("DID 必须以 'did:' 开头")

        payload = {
            "subject": {
                "$type": "com.atproto.admin.defs#repoRef",
                "did": did
            },
            "takedown": {
                "applied": False
            }
        }

        self._make_request(
            "POST",
            "com.atproto.admin.updateSubjectStatus",
            payload
        )

    def reset_password(self, did: str) -> str:
        """重置账户密码"""
        if not did.startswith("did:"):
            raise ValueError("DID 必须以 'did:' 开头")

        new_password = base64.b64encode(secrets.token_bytes(30)).decode('utf-8')
        new_password = ''.join(c for c in new_password if c.isalnum())[:24]

        self._make_request(
            "POST",
            "com.atproto.admin.updateAccountPassword",
            {
                "did": did,
                "password": new_password
            }
        )

        return new_password

def main():
    if len(sys.argv) < 2:
        print("用法: python pds_account.py <command> [args...]")
        print("可用命令:")
        print("  list")
        print("  create <email> <handle>")
        print("  delete <did>")
        print("  takedown <did>")
        print("  untakedown <did>")
        print("  reset-password <did>")
        sys.exit(1)

    command = sys.argv[1]
    pds = PDSAccount()

    try:
        if command == "list":
            accounts = pds.list_accounts()
            print("\n账户列表:")
            print("-" * 80)
            for account in accounts:
                print(f"Handle: {account.get('handle')}")
                print(f"Email: {account.get('email')}")
                print(f"DID: {account.get('did')}")
                print("-" * 80)

        elif command == "create":
            if len(sys.argv) != 4:
                print("用法: python pds_account.py create <email> <handle>")
                sys.exit(1)
            email = sys.argv[2]
            handle = sys.argv[3]
            result = pds.create_account(email, handle)
            print("\n账户创建成功!")
            print("-" * 30)
            print(f"Handle   : {handle}")
            print(f"DID      : {result['did']}")
            print(f"Password : {result['password']}")
            print("-" * 30)
            print("请保存此密码，它不会再次显示。")

        elif command == "delete":
            if len(sys.argv) != 3:
                print("用法: python pds_account.py delete <did>")
                sys.exit(1)
            did = sys.argv[2]
            response = input(f"此操作不可撤销。确定要删除 {did} 吗? [y/N] ")
            if response.lower() in ('y', 'yes'):
                pds.delete_account(did)
                print(f"{did} 已删除")

        elif command == "takedown":
            if len(sys.argv) != 3:
                print("用法: python pds_account.py takedown <did>")
                sys.exit(1)
            did = sys.argv[2]
            pds.takedown_account(did)
            print(f"{did} 已被封禁")

        elif command == "untakedown":
            if len(sys.argv) != 3:
                print("用法: python pds_account.py untakedown <did>")
                sys.exit(1)
            did = sys.argv[2]
            pds.untakedown_account(did)
            print(f"{did} 已解除封禁")

        elif command == "reset-password":
            if len(sys.argv) != 3:
                print("用法: python pds_account.py reset-password <did>")
                sys.exit(1)
            did = sys.argv[2]
            new_password = pds.reset_password(did)
            print(f"\n已重置 {did} 的密码")
            print(f"新密码: {new_password}")

        else:
            print(f"未知命令: {command}")
            sys.exit(1)

    except PDSError as e:
        print(f"错误: {str(e)}", file=sys.stderr)
        sys.exit(1)
    except requests.exceptions.RequestException as e:
        print(f"网络错误: {str(e)}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"发生错误: {str(e)}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
