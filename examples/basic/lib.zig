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
        .digest = .{ 1, 2, 3, 4 },
    };
}

pub fn login_checked(req: api.LoginRequest) api.LoginError!api.LoginResponse {
    if (rt.asSlice(req.password).len < 6) return api.LoginError.InvalidPassword;
    return .{
        .ok = true,
        .message = rt.ownString("welcome alice"),
        .token = rt.ownBytes("token-123"),
        .digest = .{ 1, 2, 3, 4 },
    };
}

pub fn rename_user(user: api.User, next_name: api.String) api.User {
    return .{
        .id = user.id,
        .kind = user.kind,
        .name = rt.ownString(rt.asSlice(next_name)),
        .email = rt.ownString(rt.asSlice(user.email)),
        .scores = user.scores,
    };
}

pub fn promote_user(user: api.User, next_kind: api.UserKind, next_scores: [3]u16) api.User {
    return .{
        .id = user.id,
        .kind = next_kind,
        .name = rt.ownString(rt.asSlice(user.name)),
        .email = rt.ownString(rt.asSlice(user.email)),
        .scores = next_scores,
    };
}

pub fn digest_name(name: api.String) api.Digest {
    const value = rt.asSlice(name);
    return .{
        if (value.len > 0) value[0] else 0,
        @as(u8, @intCast(value.len)),
        0xAB,
        0xCD,
    };
}

pub fn scale_scores(scores: api.ScoreList, factor: u16) api.ScoreList {
    const items = rt.asScoreList(scores);
    const out = std.heap.page_allocator.alloc(u16, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    for (items, 0..) |value, i| {
        out[i] = value * factor;
    }
    return rt.ownScoreList(out);
}

pub fn mirror_kind_history(history: api.UserKindList) api.UserKindList {
    const items = rt.asUserKindList(history);
    const out = std.heap.page_allocator.alloc(api.UserKind, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    @memcpy(out, items);
    return rt.ownUserKindList(out);
}

pub fn duplicate_digest(seed: api.String) api.DigestList {
    const digest = digest_name(seed);
    const out = std.heap.page_allocator.alloc([4]u8, 2) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    out[0] = digest;
    out[1] = .{ digest[0], digest[1] + 1, digest[2], digest[3] };
    return rt.ownDigestList(out);
}

pub fn mirror_metrics(metrics: api.MetricList) api.MetricList {
    const items = rt.asMetricList(metrics);
    const out = std.heap.page_allocator.alloc(api.Metric, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    @memcpy(out, items);
    return rt.ownMetricList(out);
}

pub fn mirror_users(users: api.UserList) api.UserList {
    const items = rt.asUserList(users);
    const out = std.heap.page_allocator.alloc(api.User, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    for (items, 0..) |user, i| {
        out[i] = .{
            .id = user.id,
            .kind = user.kind,
            .name = rt.ownString(rt.asSlice(user.name)),
            .email = rt.ownString(rt.asSlice(user.email)),
            .scores = user.scores,
        };
    }
    return rt.ownUserList(out);
}

pub fn mirror_buckets(buckets: api.BucketList) api.BucketList {
    const items = rt.asBucketList(buckets);
    const out = std.heap.page_allocator.alloc(api.Bucket, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    for (items, 0..) |bucket, i| {
        out[i] = .{
            .kind = bucket.kind,
            .scores = rt.ownScoreList(rt.asScoreList(bucket.scores)),
        };
    }
    return rt.ownBucketList(out);
}

pub fn maybe_kind(flag: bool) ?api.UserKind {
    if (!flag) return null;
    return api.UserKind.admin;
}

pub fn maybe_digest(flag: bool) ?api.Digest {
    if (!flag) return null;
    return .{ 9, 8, 7, 6 };
}

pub fn choose_limit(flag: bool, value: ?u32) ?u32 {
    if (!flag) return null;
    return if (value) |item| item + 1 else 1;
}

pub fn mirror_score_groups(groups: api.ScoreGroupList) api.ScoreGroupList {
    const items = rt.asScoreGroupList(groups);
    const out = std.heap.page_allocator.alloc(api.ScoreList, items.len) catch @panic("alloc failed");
    defer std.heap.page_allocator.free(out);
    for (items, 0..) |group, i| {
        out[i] = rt.ownScoreList(rt.asScoreList(group));
    }
    return rt.ownScoreGroupList(out);
}
