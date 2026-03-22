const std = @import("std");
const api = @import("api.zig");

fn allocCopy(bytes: []const u8) [*]u8 {
    const size = if (bytes.len == 0) 1 else bytes.len;
    const ptr = std.heap.c_allocator.alloc(u8, size) catch @panic("alloc failed");
    if (bytes.len > 0) {
        @memcpy(ptr[0..bytes.len], bytes);
    }
    return ptr.ptr;
}

fn asSlice(value: api.String) []const u8 {
    if (value.len == 0) return "";
    return value.ptr[0..value.len];
}

fn asBytes(value: api.Bytes) []const u8 {
    if (value.len == 0) return &.{};
    return value.ptr[0..value.len];
}

fn ownString(text: []const u8) api.String {
    return .{ .ptr = allocCopy(text), .len = text.len };
}

fn ownBytes(bytes: []const u8) api.Bytes {
    return .{ .ptr = allocCopy(bytes), .len = bytes.len };
}

pub export fn go2zig_free_buf(ptr: ?*anyopaque, len: usize) void {
    if (ptr) |value| {
        const size = if (len == 0) 1 else len;
        std.heap.c_allocator.free(@as([*]u8, @ptrCast(value))[0..size]);
    }
}

pub export fn health() bool {
    return true;
}

pub export fn login(req: api.LoginRequest) api.LoginResponse {
    const name = asSlice(req.user.name);
    const ok = name.len > 0 and asSlice(req.password).len >= 6;
    const message = if (ok)
        std.fmt.allocPrint(std.heap.c_allocator, "welcome {s}", .{name}) catch @panic("alloc failed")
    else
        std.fmt.allocPrint(std.heap.c_allocator, "invalid login for {s}", .{name}) catch @panic("alloc failed");
    defer std.heap.c_allocator.free(message);

    const token = if (ok)
        ownBytes("token-123")
    else
        ownBytes(&.{});

    return .{
        .ok = ok,
        .message = ownString(message),
        .token = token,
    };
}

pub export fn rename_user(user: api.User, next_name: api.String) api.User {
    return .{
        .id = user.id,
        .name = ownString(asSlice(next_name)),
        .email = ownString(asSlice(user.email)),
    };
}
