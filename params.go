package brotli

/**
 * Opaque structure that holds encoder state.
 *
 * Allocated and initialized with ::BrotliEncoderCreateInstance.
 * Cleaned up and deallocated with ::BrotliEncoderDestroyInstance.
 */

/**
 * Sets the specified parameter to the given encoder instance.
 *
 * @param state encoder instance
 * @param param parameter to set
 * @param value new parameter value
 * @returns ::false if parameter is unrecognized, or value is invalid
 * @returns ::false if value of parameter can not be changed at current
 *          encoder state (e.g. when encoding is started, window size might be
 *          already encoded and therefore it is impossible to change it)
 * @returns ::true if value is accepted
 * @warning invalid values might be accepted in case they would not break
 *          encoding process.
 */

/**
 * Creates an instance of ::BrotliEncoderState and initializes it.
 *
 * @p alloc_func and @p free_func @b MUST be both zero or both non-zero. In the
 * case they are both zero, default memory allocators are used. @p opaque is
 * passed to @p alloc_func and @p free_func when they are called. @p free_func
 * has to return without doing anything when asked to free a NULL pointer.
 *
 * @param alloc_func custom memory allocation function
 * @param free_func custom memory free function
 * @param opaque custom memory manager handle
 * @returns @c 0 if instance can not be allocated or initialized
 * @returns pointer to initialized ::BrotliEncoderState otherwise
 */

/**
 * Deinitializes and frees ::BrotliEncoderState instance.
 *
 * @param state decoder instance to be cleaned up and deallocated
 */

/**
 * Calculates the output size bound for the given @p input_size.
 *
 * @warning Result is only valid if quality is at least @c 2 and, in
 *          case ::BrotliEncoderCompressStream was used, no flushes
 *          (::BROTLI_OPERATION_FLUSH) were performed.
 *
 * @param input_size size of projected input
 * @returns @c 0 if result does not fit @c size_t
 */

/**
 * Performs one-shot memory-to-memory compression.
 *
 * Compresses the data in @p input_buffer into @p encoded_buffer, and sets
 * @p *encoded_size to the compressed length.
 *
 * @note If ::BrotliEncoderMaxCompressedSize(@p input_size) returns non-zero
 *       value, then output is guaranteed to be no longer than that.
 *
 * @param quality quality parameter value, e.g. ::BROTLI_DEFAULT_QUALITY
 * @param lgwin lgwin parameter value, e.g. ::BROTLI_DEFAULT_WINDOW
 * @param mode mode parameter value, e.g. ::BROTLI_DEFAULT_MODE
 * @param input_size size of @p input_buffer
 * @param input_buffer input data buffer with at least @p input_size
 *        addressable bytes
 * @param[in, out] encoded_size @b in: size of @p encoded_buffer; \n
 *                 @b out: length of compressed data written to
 *                 @p encoded_buffer, or @c 0 if compression fails
 * @param encoded_buffer compressed data destination buffer
 * @returns ::false in case of compression error
 * @returns ::false if output buffer is too small
 * @returns ::true otherwise
 */

/**
 * Compresses input stream to output stream.
 *
 * The values @p *available_in and @p *available_out must specify the number of
 * bytes addressable at @p *next_in and @p *next_out respectively.
 * When @p *available_out is @c 0, @p next_out is allowed to be @c NULL.
 *
 * After each call, @p *available_in will be decremented by the amount of input
 * bytes consumed, and the @p *next_in pointer will be incremented by that
 * amount. Similarly, @p *available_out will be decremented by the amount of
 * output bytes written, and the @p *next_out pointer will be incremented by
 * that amount.
 *
 * @p total_out, if it is not a null-pointer, will be set to the number
 * of bytes compressed since the last @p state initialization.
 *
 *
 *
 * Internally workflow consists of 3 tasks:
 *  -# (optionally) copy input data to internal buffer
 *  -# actually compress data and (optionally) store it to internal buffer
 *  -# (optionally) copy compressed bytes from internal buffer to output stream
 *
 * Whenever all 3 tasks can't move forward anymore, or error occurs, this
 * method returns the control flow to caller.
 *
 * @p op is used to perform flush, finish the stream, or inject metadata block.
 * See ::BrotliEncoderOperation for more information.
 *
 * Flushing the stream means forcing encoding of all input passed to encoder and
 * completing the current output block, so it could be fully decoded by stream
 * decoder. To perform flush set @p op to ::BROTLI_OPERATION_FLUSH.
 * Under some circumstances (e.g. lack of output stream capacity) this operation
 * would require several calls to ::BrotliEncoderCompressStream. The method must
 * be called again until both input stream is depleted and encoder has no more
 * output (see ::BrotliEncoderHasMoreOutput) after the method is called.
 *
 * Finishing the stream means encoding of all input passed to encoder and
 * adding specific "final" marks, so stream decoder could determine that stream
 * is complete. To perform finish set @p op to ::BROTLI_OPERATION_FINISH.
 * Under some circumstances (e.g. lack of output stream capacity) this operation
 * would require several calls to ::BrotliEncoderCompressStream. The method must
 * be called again until both input stream is depleted and encoder has no more
 * output (see ::BrotliEncoderHasMoreOutput) after the method is called.
 *
 * @warning When flushing and finishing, @p op should not change until operation
 *          is complete; input stream should not be swapped, reduced or
 *          extended as well.
 *
 * @param state encoder instance
 * @param op requested operation
 * @param[in, out] available_in @b in: amount of available input; \n
 *                 @b out: amount of unused input
 * @param[in, out] next_in pointer to the next input byte
 * @param[in, out] available_out @b in: length of output buffer; \n
 *                 @b out: remaining size of output buffer
 * @param[in, out] next_out compressed output buffer cursor;
 *                 can be @c NULL if @p available_out is @c 0
 * @param[out] total_out number of bytes produced so far; can be @c NULL
 * @returns ::false if there was an error
 * @returns ::true otherwise
 */

/**
 * Checks if encoder instance reached the final state.
 *
 * @param state encoder instance
 * @returns ::true if encoder is in a state where it reached the end of
 *          the input and produced all of the output
 * @returns ::false otherwise
 */

/**
 * Checks if encoder has more output.
 *
 * @param state encoder instance
 * @returns ::true, if encoder has some unconsumed output
 * @returns ::false otherwise
 */

/**
 * Acquires pointer to internal output buffer.
 *
 * This method is used to make language bindings easier and more efficient:
 *  -# push data to ::BrotliEncoderCompressStream,
 *     until ::BrotliEncoderHasMoreOutput returns BROTL_TRUE
 *  -# use ::BrotliEncoderTakeOutput to peek bytes and copy to language-specific
 *     entity
 *
 * Also this could be useful if there is an output stream that is able to
 * consume all the provided data (e.g. when data is saved to file system).
 *
 * @attention After every call to ::BrotliEncoderTakeOutput @p *size bytes of
 *            output are considered consumed for all consecutive calls to the
 *            instance methods; returned pointer becomes invalidated as well.
 *
 * @note Encoder output is not guaranteed to be contiguous. This means that
 *       after the size-unrestricted call to ::BrotliEncoderTakeOutput,
 *       immediate next call to ::BrotliEncoderTakeOutput may return more data.
 *
 * @param state encoder instance
 * @param[in, out] size @b in: number of bytes caller is ready to take, @c 0 if
 *                 any amount could be handled; \n
 *                 @b out: amount of data pointed by returned pointer and
 *                 considered consumed; \n
 *                 out value is never greater than in value, unless it is @c 0
 * @returns pointer to output data
 */

/**
 * Gets an encoder library version.
 *
 * Look at BROTLI_VERSION for more information.
 */
/* Copyright 2017 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Parameters for the Brotli encoder with chosen quality levels. */
type hasherParams struct {
	type_                       int
	bucket_bits                 int
	block_bits                  int
	hash_len                    int
	num_last_distances_to_check int
}

type distanceParams struct {
	distance_postfix_bits     uint32
	num_direct_distance_codes uint32
	alphabet_size             uint32
	max_distance              uint
}

/* Encoding parameters */
type encoderParams struct {
	mode                             int
	quality                          int
	lgwin                            uint
	lgblock                          int
	size_hint                        uint
	disable_literal_context_modeling bool
	large_window                     bool
	hasher                           hasherParams
	dist                             distanceParams
	dictionary                       encoderDictionary
}
