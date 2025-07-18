// Ninstyle sound effects for task manager
"use strict";

(function(window) {
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    let audioContext = null;
    let soundEnabled = false; // Sounds disabled by default
    
    // Initialize audio context on first user interaction
    function initAudio() {
        if (!audioContext && soundEnabled) {
            audioContext = new AudioContext();
        }
    }
    
    // Toggle sound on/off
    function toggleSound() {
        soundEnabled = !soundEnabled;
        localStorage.setItem('done_soundEnabled', JSON.stringify(soundEnabled));
        if (soundEnabled) {
            initAudio();
        }
        return soundEnabled;
    }
    
    // Load sound preference
    function loadSoundPreference() {
        const stored = localStorage.getItem('done_soundEnabled');
        if (stored !== null) {
            soundEnabled = JSON.parse(stored);
        }
    }
    
    // Play a simple beep sound
    function playBeep(frequency, duration, volume = 0.3) {
        if (!soundEnabled) return;
        initAudio();
        if (!audioContext) return;
        
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();
        
        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);
        
        oscillator.frequency.value = frequency;
        oscillator.type = 'square';
        
        gainNode.gain.setValueAtTime(volume, audioContext.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + duration);
        
        oscillator.start(audioContext.currentTime);
        oscillator.stop(audioContext.currentTime + duration);
    }
    
    // Ninstyle sound effects
    const sounds = {
        // Task creation sound - rising tones like Mario coin
        taskCreate: function() {
            playBeep(523.25, 0.1); // C5
            setTimeout(() => playBeep(659.25, 0.1), 50); // E5
            setTimeout(() => playBeep(783.99, 0.15), 100); // G5
        },
        
        // Task start sound - power up
        taskStart: function() {
            playBeep(261.63, 0.1); // C4
            setTimeout(() => playBeep(329.63, 0.1), 100); // E4
            setTimeout(() => playBeep(392.00, 0.1), 200); // G4
            setTimeout(() => playBeep(523.25, 0.2), 300); // C5
        },
        
        // Task complete sound - victory fanfare
        taskComplete: function() {
            playBeep(523.25, 0.1); // C5
            setTimeout(() => playBeep(523.25, 0.1), 100); // C5
            setTimeout(() => playBeep(523.25, 0.1), 200); // C5
            setTimeout(() => playBeep(659.25, 0.3), 300); // E5
        },
        
        // Task delete sound - descending tones
        taskDelete: function() {
            playBeep(392.00, 0.1); // G4
            setTimeout(() => playBeep(329.63, 0.1), 50); // E4
            setTimeout(() => playBeep(261.63, 0.2), 100); // C4
        },
        
        // Button hover sound - soft click
        buttonHover: function() {
            playBeep(880.00, 0.05, 0.1); // A5
        },
        
        // Drag start sound
        dragStart: function() {
            playBeep(440.00, 0.1); // A4
            setTimeout(() => playBeep(554.37, 0.1), 50); // C#5
        },
        
        // Drop sound
        drop: function() {
            playBeep(554.37, 0.1); // C#5
            setTimeout(() => playBeep(440.00, 0.1), 50); // A4
        }
    };
    
    // Export sounds and controls
    window.NinstyleSounds = {
        ...sounds,
        toggleSound: toggleSound,
        isEnabled: () => soundEnabled,
        init: loadSoundPreference
    };
    
    // Initialize audio on first user interaction
    document.addEventListener('click', initAudio, { once: true });
    document.addEventListener('touchstart', initAudio, { once: true });
    
    // Load sound preference on startup
    loadSoundPreference();
    
})(window);