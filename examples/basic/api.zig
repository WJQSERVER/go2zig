pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Digest = [4]u8;

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const ScoreList = extern struct {
    ptr: ?[*]const u16,
    len: usize,
};

pub const UserKindList = extern struct {
    ptr: ?[*]const UserKind,
    len: usize,
};

pub const DigestList = extern struct {
    ptr: ?[*]const Digest,
    len: usize,
};

pub const MetricList = extern struct {
    ptr: ?[*]const Metric,
    len: usize,
};

pub const UserList = extern struct {
    ptr: ?[*]const User,
    len: usize,
};

pub const BucketList = extern struct {
    ptr: ?[*]const Bucket,
    len: usize,
};

pub const UserKind = enum(u8) {
    guest,
    member,
    admin,
};

pub const User = extern struct {
    id: u64,
    kind: UserKind,
    name: String,
    email: String,
    scores: [3]u16,
};

pub const Metric = extern struct {
    kind: UserKind,
    scores: [3]u16,
};

pub const Bucket = extern struct {
    kind: UserKind,
    scores: ScoreList,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
    digest: Digest,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn health() bool;
pub extern fn login(req: LoginRequest) LoginResponse;
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
pub extern fn rename_user(user: User, next_name: String) User;
pub extern fn promote_user(user: User, next_kind: UserKind, next_scores: [3]u16) User;
pub extern fn digest_name(name: String) Digest;
pub extern fn scale_scores(scores: ScoreList, factor: u16) ScoreList;
pub extern fn mirror_kind_history(history: UserKindList) UserKindList;
pub extern fn duplicate_digest(seed: String) DigestList;
pub extern fn mirror_metrics(metrics: MetricList) MetricList;
pub extern fn mirror_users(users: UserList) UserList;
pub extern fn mirror_buckets(buckets: BucketList) BucketList;
pub extern fn maybe_kind(flag: bool) ?UserKind;
pub extern fn maybe_digest(flag: bool) ?Digest;
pub extern fn choose_limit(flag: bool, value: ?u32) ?u32;
