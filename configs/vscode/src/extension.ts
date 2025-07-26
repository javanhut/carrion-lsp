import * as vscode from 'vscode';
import * as path from 'path';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind,
    ExecutableOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('Carrion Language Support extension is now active!');

    // Register commands
    registerCommands(context);

    // Start the language server
    startLanguageServer(context);

    // Setup format on save
    setupFormatOnSave(context);
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

function registerCommands(context: vscode.ExtensionContext) {
    // Restart server command
    const restartServerCommand = vscode.commands.registerCommand('carrion.restartServer', async () => {
        if (client) {
            await client.stop();
            startLanguageServer(context);
            vscode.window.showInformationMessage('Carrion language server restarted');
        }
    });

    // Show output channel command
    const showOutputCommand = vscode.commands.registerCommand('carrion.showOutputChannel', () => {
        if (client) {
            client.outputChannel.show();
        }
    });

    context.subscriptions.push(restartServerCommand, showOutputCommand);
}

function startLanguageServer(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('carrion');
    
    // Check if the language server is enabled
    if (!config.get('enable', true)) {
        console.log('Carrion language server is disabled');
        return;
    }

    // Get server path from configuration
    const serverPath = config.get('serverPath', 'carrion-lsp') as string;
    
    // Server options
    const serverOptions: ServerOptions = {
        command: serverPath,
        args: [],
        options: {
            env: {
                ...process.env,
            }
        } as ExecutableOptions,
    };

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        // Register the server for Carrion documents
        documentSelector: [
            { scheme: 'file', language: 'carrion' },
            { scheme: 'untitled', language: 'carrion' }
        ],
        
        // Synchronize settings
        synchronize: {
            // Notify the server about file changes to '.crl files contained in the workspace
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.crl')
        },

        // Pass the configuration to the server
        initializationOptions: {
            settings: {
                carrion: {
                    diagnostics: {
                        enable: config.get('diagnostics.enable', true),
                    },
                    completion: {
                        enable: config.get('completion.enable', true),
                        memberAccess: config.get('completion.memberAccess', true),
                    },
                    formatting: {
                        enable: config.get('formatting.enable', true),
                    },
                },
            },
        },

        // Output channel
        outputChannelName: 'Carrion Language Server',
        
        // Trace setting
        traceOutputChannel: vscode.window.createOutputChannel('Carrion Language Server Trace'),
    };

    // Create the language client and start the client
    client = new LanguageClient(
        'carrionLanguageServer',
        'Carrion Language Server',
        serverOptions,
        clientOptions
    );

    // Set trace level from configuration
    const traceLevel = config.get('trace.server', 'off') as string;
    client.setTrace(traceLevel as any);

    // Start the client. This will also launch the server
    client.start().then(() => {
        console.log('Carrion language server started successfully');
        
        // Show notification on first activation
        const hasShownWelcome = context.globalState.get('carrion.hasShownWelcome', false);
        if (!hasShownWelcome) {
            vscode.window.showInformationMessage(
                'Carrion Language Support is now active! Enjoy coding in Carrion.',
                'Learn More'
            ).then(selection => {
                if (selection === 'Learn More') {
                    vscode.env.openExternal(vscode.Uri.parse('https://github.com/javanhut/carrion-lsp'));
                }
            });
            context.globalState.update('carrion.hasShownWelcome', true);
        }
    }).catch(error => {
        console.error('Failed to start Carrion language server:', error);
        vscode.window.showErrorMessage(
            `Failed to start Carrion language server: ${error.message}. ` +
            'Make sure carrion-lsp is installed and in your PATH.'
        );
    });

    context.subscriptions.push(client);
}

function setupFormatOnSave(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('carrion');
    
    if (config.get('formatting.formatOnSave', true)) {
        const formatOnSave = vscode.workspace.onWillSaveTextDocument(async (event) => {
            const document = event.document;
            
            // Only format Carrion files
            if (document.languageId !== 'carrion') {
                return;
            }

            // Check if formatting is enabled
            const currentConfig = vscode.workspace.getConfiguration('carrion');
            if (!currentConfig.get('formatting.enable', true)) {
                return;
            }

            // Format the document
            event.waitUntil(
                vscode.commands.executeCommand('editor.action.formatDocument')
            );
        });

        context.subscriptions.push(formatOnSave);
    }
}

// Helper function to check if carrion-lsp is available
async function checkServerAvailability(): Promise<boolean> {
    const config = vscode.workspace.getConfiguration('carrion');
    const serverPath = config.get('serverPath', 'carrion-lsp') as string;
    
    try {
        const { exec } = require('child_process');
        return new Promise((resolve) => {
            exec(`${serverPath} --version`, (error: any) => {
                resolve(!error);
            });
        });
    } catch {
        return false;
    }
}