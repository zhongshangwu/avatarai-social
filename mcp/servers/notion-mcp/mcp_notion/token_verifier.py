import logging

from mcp.server.auth.provider import AccessToken, TokenVerifier

logger = logging.getLogger(__name__)


class DummyTokenVerifier(TokenVerifier):

    async def verify_token(self, token: str) -> AccessToken | None:
        return AccessToken(
            token=token,
            client_id="dummy",
            scopes=["dummy"],
        )
