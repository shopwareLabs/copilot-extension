import { readFileSync } from 'node:fs';

const output = readFileSync('data/commands.json', 'utf8');

const uselessCommands = ['help', 'version', 'exit', '_complete', 'clear', 'history', 'about', 'sync:composer:version', 'secrets:encrypt-from-local', 'secrets:decrypt-to-local', 'secrets:generate-keys', 'secrets:list', 'secrets:remove', 'secrets:reveal', 'secrets:set', 's3:set-visibility', 'completion'];

for (const command of JSON.parse(output).commands) {
    if (uselessCommands.includes(command.name)) {
        continue;
    }

    console.log(`# ${command.usage}`);
    console.log('')

    if (command.description) {
        console.log(command.description);
    }

    if (Object.keys(command.definition.options)) {
        console.log('')
        console.log('## Options')
        console.log('')

        for (const [name, option] of Object.entries(command.definition.options)) {
            if (['help', 'version', 'silent', 'verbose', 'quiet', 'ansi', 'no-ansi', 'no-interaction', 'profile', 'no-debug', 'env'].includes(name)) {
                continue;
            }

            console.log(`- \`--${name}${option.shortcut ? `|${option.shortcut}` : ''}\` - ${option.description}${option.default ? ` (default: ${option.default})` : ''}`);
        }
    }

    console.log('')
}