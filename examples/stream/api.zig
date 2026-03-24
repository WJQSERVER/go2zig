pub const GoReader = usize;
pub const GoWriter = usize;

pub extern fn copy_stream(reader: GoReader, writer: GoWriter) u64;
