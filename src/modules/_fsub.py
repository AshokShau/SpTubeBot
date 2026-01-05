from typing import TypeAlias, Union, Callable, Awaitable
from functools import wraps
from pytdbot import Client, types
from src.config import FSUB_ID

ChatMemberStatus: TypeAlias = Union[
    types.ChatMemberStatusCreator,
    types.ChatMemberStatusAdministrator,
    types.ChatMemberStatusMember,
    types.ChatMemberStatusRestricted,
    types.ChatMemberStatusLeft,
    types.ChatMemberStatusBanned,
]

BLOCKED_STATUSES = {
    types.ChatMemberStatusLeft().getType(),
    types.ChatMemberStatusBanned().getType(),
    types.ChatMemberStatusRestricted().getType(),
}

# Caches
member_status_cache = {}
invite_link_cache: dict[int, str] = {}

def fsub(func: Callable[..., Awaitable]):
    @wraps(func)
    async def wrapper(client: Client, message: types.Message, *args, **kwargs):
        chat_id = message.chat_id
        if chat_id < 0:
            return await func(client, message, *args, **kwargs)

        # FSUB disabled
        if not FSUB_ID or FSUB_ID == 0:
            return await func(client, message, *args, **kwargs)

        user_id = message.from_id
        cached_status = member_status_cache.get(user_id)

        # If user is already verified, skip FSUB check
        if cached_status and cached_status not in BLOCKED_STATUSES:
            return await func(client, message, *args, **kwargs)

        # Get member status
        member = await client.getChatMember(
            chat_id=FSUB_ID,
            member_id=types.MessageSenderUser(user_id)
        )
        if isinstance(member, types.Error) or member.status is None:
            if member.code == 400 and member.message == "Chat not found":
                client.logger.warning(f"FSUB group not found: {FSUB_ID}")
                return await func(client, message, *args, **kwargs)
            status_type = types.ChatMemberStatusLeft().getType()
        else:
            status_type = member.status.getType()

        # Save verified users in cache
        if status_type not in BLOCKED_STATUSES:
            member_status_cache[user_id] = status_type
            return await func(client, message, *args, **kwargs)

        # Get invite link from cache or TG
        invite_link = invite_link_cache.get(FSUB_ID)
        if not invite_link:
            _chat_id = int(str(FSUB_ID)[4:]) if str(FSUB_ID).startswith("-100") else FSUB_ID
            chat_info = await client.getSupergroupFullInfo(_chat_id)
            if isinstance(chat_info, types.Error):
                client.logger.warning(f"❌ Failed to get supergroup info: {chat_info.message}")
                return await func(client, message, *args, **kwargs)

            invite_link = getattr(chat_info.invite_link, "invite_link", None)
            if invite_link:
                invite_link_cache[FSUB_ID] = invite_link
            else:
                client.logger.warning(f"❌ No invite link found for: {FSUB_ID}")
                return await func(client, message, *args, **kwargs)

        # Send FSUB message
        text = (
            "🔒 <b>Channel Membership Required</b>\n\n"
            "You need to join our channel to use me in private chat.\n"
            "✅ Once you’ve joined, type /start here to activate me.\n\n"
            "💬 In groups, you can use me without a subscription."
        )
        button = types.ReplyMarkupInlineKeyboard(
            [[types.InlineKeyboardButton(text="📢 Join Channel",
                                         type=types.InlineKeyboardButtonTypeUrl(url=invite_link))]]
        )

        return await message.reply_text(
            text=text,
            parse_mode="html",
            disable_web_page_preview=True,
            reply_markup=button,
        )

    return wrapper


def is_valid_supergroup(chat_id: int) -> bool:
    """
    Check if a chat ID is for a supergroup.
    """
    return str(chat_id).startswith("-100")

async def _validate_chat(chat_id: int) -> bool:
    """Validate if chat is a supergroup and handle non-supergroups."""
    return bool(is_valid_supergroup(chat_id))

@Client.on_updateChatMember()
async def chat_member(client: Client, update: types.UpdateChatMember) -> None:
    """Handles member updates in the chat (joins, leaves, promotions, etc.)."""
    chat_id = update.chat_id

    # Early return for non-group chats
    if chat_id > 0 or not await _validate_chat(chat_id):
        return None

    if FSUB_ID != chat_id:
        return None

    user_id = update.new_chat_member.member_id.user_id
    old_status = update.old_chat_member.status["@type"]
    new_status = update.new_chat_member.status["@type"]

    # Handle different status change scenarios
    await _handle_status_changes(client, chat_id, user_id, old_status, new_status)
    return None

async def _handle_status_changes(
    client: Client, chat_id: int, user_id: int, old_status: str, new_status: str
) -> None:
    """Route different status change scenarios to appropriate handlers."""
    if old_status == "chatMemberStatusLeft" and new_status in {
        "chatMemberStatusMember",
        "chatMemberStatusAdministrator",
    }:
        await _handle_join(client, chat_id, user_id)
    elif (
        old_status in {"chatMemberStatusMember", "chatMemberStatusAdministrator"}
        and new_status == "chatMemberStatusLeft"
    ):
        await _handle_leave_or_kick(client, chat_id, user_id)
    elif new_status == "chatMemberStatusBanned":
        await _handle_ban(chat_id, user_id)
    elif (
        old_status == "chatMemberStatusBanned" and new_status == "chatMemberStatusLeft"
    ):
        await _handle_unban(chat_id, user_id)

async def _handle_join(client: Client, chat_id: int, user_id: int) -> None:
    """Handle user/bot joining the chat."""
    member_status_cache[user_id] = types.ChatMemberStatusMember().getType()

async def _handle_leave_or_kick(client: Client, chat_id: int, user_id: int) -> None:
    """Handle user leaving or being kicked from chat."""
    member_status_cache[user_id] = types.ChatMemberStatusLeft().getType()

async def _handle_ban(chat_id: int, user_id: int) -> None:
    """Handle user being banned from chat."""
    member_status_cache[user_id] = types.ChatMemberStatusBanned().getType()


async def _handle_unban(chat_id: int, user_id: int) -> None:
    """Handle user being unbanned from chat."""
    member_status_cache[user_id] = types.ChatMemberStatusLeft().getType()
