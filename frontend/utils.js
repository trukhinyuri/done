/**
 * Utility functions for encoding/decoding special characters
 * Uses base64 encoding instead of hardcoded UUIDs for better security
 */
"use strict";

(function(window) {
    
    /**
     * Encode special characters in task text for safe transmission
     * @param {string} text - The text to encode
     * @returns {string} - Encoded text
     */
    function encodeTaskText(text) {
        if (!text) return text;
        
        // Use URL-safe base64 encoding
        try {
            // First escape unicode properly
            var utf8 = unescape(encodeURIComponent(text));
            var base64 = btoa(utf8);
            // Make it URL-safe
            return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
        } catch (e) {
            console.error('Failed to encode task text:', e);
            // Fallback to simple escaping
            return text
                .replace(/\\/g, '\\\\')
                .replace(/"/g, '\\"')
                .replace(/'/g, "\\'")
                .replace(/\n/g, '\\n')
                .replace(/\r/g, '\\r')
                .replace(/\t/g, '\\t');
        }
    }
    
    /**
     * Decode special characters in task text
     * @param {string} encodedText - The encoded text
     * @returns {string} - Decoded text
     */
    function decodeTaskText(encodedText) {
        if (!encodedText) return encodedText;
        
        try {
            // Check if it's base64 encoded (doesn't contain spaces or special chars except - and _)
            if (/^[A-Za-z0-9\-_]+$/.test(encodedText)) {
                // Restore base64 padding
                var base64 = encodedText.replace(/-/g, '+').replace(/_/g, '/');
                var pad = base64.length % 4;
                if (pad) {
                    base64 += '===='.substring(pad);
                }
                var utf8 = atob(base64);
                return decodeURIComponent(escape(utf8));
            }
        } catch (e) {
            // Not base64 encoded, might be legacy UUID format
        }
        
        // Handle legacy UUID format for backward compatibility
        if (encodedText.indexOf('280d382c-f23e-4631-8551-f43661405497') !== -1 ||
            encodedText.indexOf('e6f23f57-6cad-451b-8306-7939e25542dc') !== -1 ||
            encodedText.indexOf('a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f') !== -1) {
            return encodedText
                .split('280d382c-f23e-4631-8551-f43661405497').join('"')
                .split('e6f23f57-6cad-451b-8306-7939e25542dc').join("'")
                .split('a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f').join('\n');
        }
        
        // Handle simple escaped format
        if (encodedText.indexOf('\\') !== -1) {
            return encodedText
                .replace(/\\n/g, '\n')
                .replace(/\\r/g, '\r')
                .replace(/\\t/g, '\t')
                .replace(/\\'/g, "'")
                .replace(/\\"/g, '"')
                .replace(/\\\\/g, '\\');
        }
        
        // Return as-is if no encoding detected
        return encodedText;
    }
    
    /**
     * Clean task text for display (remove any encoding artifacts)
     * @param {string} text - The text to clean
     * @returns {string} - Clean text
     */
    function cleanTaskText(text) {
        if (!text) return text;
        
        // First try to decode
        var decoded = decodeTaskText(text);
        
        // Remove any remaining UUID artifacts (in case of double encoding)
        decoded = decoded
            .split('280d382c-f23e-4631-8551-f43661405497').join('"')
            .split('e6f23f57-6cad-451b-8306-7939e25542dc').join("'")
            .split('a7f3d0a1-2b5e-4c6d-8e9f-1a2b3c4d5e6f').join('\n');
            
        return decoded;
    }
    
    // Export functions
    window.TaskUtils = {
        encodeTaskText: encodeTaskText,
        decodeTaskText: decodeTaskText,
        cleanTaskText: cleanTaskText
    };
    
})(window);