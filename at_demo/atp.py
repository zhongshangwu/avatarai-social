from atproto import Client, Request
from atproto_client import models
import os
import fire
# --- 登录凭据 ---
# 强烈建议使用环境变量来存储敏感信息，而不是直接硬编码在代码中
# 你也可以直接替换下面的字符串为你自己的邮箱和应用密码
# 重要提示：请使用你的 Bluesky 应用密码 (App Password)，而不是你的主账户密码！
BLUESKY_EMAIL = os.environ.get("BLUESKY_EMAIL", "ouyzam7.pds.avatar.ai")
BLUESKY_PASSWORD = os.environ.get("BLUESKY_PASSWORD", "iPhone520.") # 格式通常是 xxxx-xxxx-xxxx-xxxx


def test():
    # --- 登录并获取会话 ---
    try:
        request = Request(verify=False)
        # 1. 创建 Client 实例
        client = Client(base_url="https://pds.avatar.ai", request=request)

        # 2. 调用 login 方法进行登录
        # 该方法会尝试登录，并在成功后将认证信息存储在 client 实例中
        # login 方法本身也会返回会话信息字典
        print(f"尝试使用邮箱 {BLUESKY_EMAIL} 登录...")
        session_info = client.com.atproto.server.create_session(models.ComAtprotoServerCreateSession.Data(identifier=BLUESKY_EMAIL, password=BLUESKY_PASSWORD))

        print("\n登录成功!")
        print("------ 通过 login() 返回的会话信息 ------")
        print(f"用户 DID: {session_info}")
        print(f"用户 Handle: {session_info.handle}")
        # 注意：如果登录时使用的是 handle 而不是 email，这里可能为 None
        print(f"用户 Email: {session_info.email}")
        print(f"Access Token (JWT): {session_info.access_jwt[:15]}...") # 只显示部分 Token
        print(f"Refresh Token (JWT): {session_info.refresh_jwt[:15]}...") # 只显示部分 Token

        # 3. 你也可以在登录后，随时通过 client 的属性访问会话信息
        # (这在你需要在代码的其他地方访问时很有用)
        print("\n------ 通过 client 实例访问的当前会话 ------")
        profile = client.login(BLUESKY_EMAIL, BLUESKY_PASSWORD)

        print(profile)

        # 获取用户 profile 详细信息
        try:
            profile = client.app.bsky.actor.get_profile(params=models.AppBskyActorGetProfile.Params(actor=session_info.did))

            print("\n------ 用户 Profile 信息 ------")
            print(f"显示名称: {profile.display_name}")
            print(f"用户描述: {profile.description}")
            print(f"关注人数: {profile.follows_count}")
            print(f"粉丝人数: {profile.followers_count}")
            print(f"发帖数量: {profile.posts_count}")

            # 如果有头像信息
            if hasattr(profile, 'avatar'):
                print(f"头像链接: {profile.avatar}")

            # 如果有背景图片
            if hasattr(profile, 'banner'):
                print(f"背景图片: {profile.banner}")

        except Exception as e:
            print(f"获取用户 profile 时发生错误: {e}")


        # 通过原始的 profile record 获取头像和背景图片
        profile_record = client.com.atproto.repo.get_record(params=models.ComAtprotoRepoGetRecord.Params(
            repo=session_info.did,
            collection="app.bsky.actor.profile", rkey="self"))
        print(profile_record)

        # 更新 profile record, 修改显示名称
        try:
            # 首先获取现有的 profile record
            profile_record = client.com.atproto.repo.get_record(params=models.ComAtprotoRepoGetRecord.Params(
                repo=session_info.did,
                collection="app.bsky.actor.profile",
                rkey="self"
            )).value

            print(profile_record)

            raise Exception("test")

            # 准备新的 profile 数据，保留原有字段
            new_profile_data = {
                "$type": "app.bsky.actor.profile",
                "displayName": "新的显示名称",  # 修改为你想要的显示名称
                "description": "这是新的个人简介",  # 修改为你想要的简介
                "avatar": profile_record.avatar,  # 保持原有头像
                "created_at": profile_record.created_at,  # 保持原有创建时间
            }

            # 如果原记录有其他非空字段，也添加到新数据中
            if profile_record.banner:
                new_profile_data["banner"] = profile_record.banner
            if profile_record.labels:
                new_profile_data["labels"] = profile_record.labels
            if profile_record.joined_via_starter_pack:
                new_profile_data["joined_via_starter_pack"] = profile_record.joined_via_starter_pack
            if profile_record.pinned_post:
                new_profile_data["pinned_post"] = profile_record.pinned_post

            # 使用 put_record 更新 profile
            updated_profile = client.com.atproto.repo.put_record(models.ComAtprotoRepoPutRecord.Data(
                repo=session_info.did,
                collection="app.bsky.actor.profile",
                rkey="self",
                record=new_profile_data
            ))

            print("\n------ Profile 更新成功 ------")
            print(f"更新后的记录: {updated_profile}")

        except Exception as e:
            print(f"更新 profile 时发生错误: {e}")


        # 现在 client 对象已经认证，你可以用它进行其他操作，例如发帖
        # response = client.send_post(text='使用 atproto 库通过 Python 登录成功！')
        # print(f"\n发帖成功: {response.uri}")

    except Exception as e:
        print(f"\n登录或操作时发生错误: {e}")
        print("请检查你的邮箱和应用密码是否正确，以及网络连接是否正常。")
        raise

def login_client():
    request = Request(verify=False)
    client = Client(base_url="https://pds.avatar.ai", request=request)
    profile = client.login(BLUESKY_EMAIL, BLUESKY_PASSWORD)
    print(profile)
    return client, profile


def get_avatar_profile_record():
    client, profile = login_client()
    profile_record = client.com.atproto.repo.get_record(params=models.ComAtprotoRepoGetRecord.Params(
        repo=profile["did"],
        collection="app.vtri.avatar.profile", rkey="self"))
    print(profile_record)


def get_posts_detail():
    client, profile = login_client()
    response = client.app.bsky.feed.get_posts(params=models.AppBskyFeedGetPosts.Params(
        uris=["at://did:plc:kvwvcn5iqfooopmyzvb4qzba/app.bsky.feed.post/3k23k23k23k23k23k23k23k23k23k23"]
    ))
    print(response)


if __name__ == "__main__":
    fire.Fire()