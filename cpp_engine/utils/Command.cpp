#include "Command.h"
#include "Log.h"

#include <vector>
#include <sstream>
#include <iostream>


const std::vector<std::string> ValidCommands = {
    "EXIT", "START", "STOP", "WORKER", "RECORDING", "NORECORDING", "PROBE"
};

std::queue<std::string> strsplit(const std::string& str, char delimiter) {
    std::queue<std::string> tokens;
    std::string token;
    std::istringstream tokenStream(str);

    while (std::getline(tokenStream, token, delimiter)) {
        tokens.push(token);
    }
    return tokens;
}

Command parseCommand(std::string line) {
    std::queue<std::string> tokens = strsplit(line, ' ');

    // ERROR 3 FIX: Allow commands with no args (size == 1)
    if (!tokens.empty()) { 

        // ERROR 2 FIX: Queue uses .front(), not [0]
        std::string name = tokens.front(); 
        
        // Now we can iterate over the vector
        for (const std::string& cmd : ValidCommands) {
            if (name == cmd) {
                tokens.pop(); // Remove the command name from the queue

                // Return the name and the remaining queue (arguments)
                return {name, tokens}; 
            }
        }
    }

    Log::info("No valid command");

    // Return empty command if not valid
    return {{}, std::queue<std::string>()}; 
}