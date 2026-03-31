pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const GoReader = usize;
pub const GoWriter = usize;

pub const CopyStreamError = error{ StreamReadFailed, StreamWriteFailed };

pub extern fn copy_stream(reader: GoReader, writer: GoWriter) CopyStreamError!u64;
