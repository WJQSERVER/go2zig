const std = @import("std");
const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn health() bool {
    return true;
}

pub fn login(req: api.LoginRequest) api.LoginResponse {
    const name = rt.asSlice(req.user.name);
    const ok = name.len > 0 and rt.asSlice(req.password).len >= 6;
    const message = if (ok)
        std.fmt.allocPrint(std.heap.page_allocator, "welcome {s}", .{name}) catch @panic("alloc failed")
    else
        std.fmt.allocPrint(std.heap.page_allocator, "invalid login for {s}", .{name}) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(message);

    return .{
        .ok = ok,
        .message = rt.ownString(message),
        .token = if (ok) rt.ownBytes("token-123") else rt.ownBytes(&.{}),
    };
}

pub fn login_checked(req: api.LoginRequest) api.LoginError!api.LoginResponse {
    if (rt.asSlice(req.password).len < 6) return api.LoginError.InvalidPassword;
    return .{
        .ok = true,
        .message = rt.ownString("welcome alice"),
        .token = rt.ownBytes("token-123"),
    };
}

pub fn rename_user(user: api.User, next_name: api.String) api.User {
    return .{
        .id = user.id,
        .name = rt.ownString(rt.asSlice(next_name)),
        .email = rt.ownString(rt.asSlice(user.email)),
    };
}
