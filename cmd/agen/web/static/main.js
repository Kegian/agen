// Elements
var buttons = {
    "swagger": document.getElementById('swagger-tab'),
    "openapi": document.getElementById('openapi-tab'),
    "youtrack": document.getElementById('youtrack-tab'),
};

mainTab = document.getElementById('main-tab');
logsTab = document.getElementById('logs-tab');

generateTab = document.getElementById('generate-tab');
saveTab = document.getElementById('save-tab');
pathTab = document.getElementById('path-tab');

editorOAPITab = document.getElementById('editor-oapi');
editorSwaggerTab = document.getElementById('editor-swagger');
editorYoutrackTab = document.getElementById('editor-youtrack');

var generated = false

// Initialize highlight editor
var editor = ace.edit("editor");
editor.setTheme("ace/theme/monokai");
editor.session.setMode("ace/mode/yaml");
editor.setOptions({
    useSoftTabs: true,
    tabSize: 2,
    enableBasicAutocompletion: true,
    showPrintMargin: false,
    fixedWidthGutter: true,
    scrollPastEnd: 0.5,
    displayIndentGuides: true,
    fontSize: "medium",
}); // https://github.com/ajaxorg/ace/wiki/Configuring-Ace
editor.session.on('change', function() {
    showGenerate(true);
})

var editorOAPI = ace.edit("editor-oapi");
editorOAPI.setTheme("ace/theme/monokai");
editorOAPI.session.setMode("ace/mode/yaml");
editorOAPI.setOptions({
    useSoftTabs: true,
    tabSize: 2,
    enableBasicAutocompletion: true,
    showPrintMargin: false,
    fixedWidthGutter: true,
    scrollPastEnd: 0.5,
    displayIndentGuides: true,
    fontSize: "medium",
    readOnly: true,
});
editorOAPITab.style.height = '100%';
editorOAPITab.style.width = '100%';

var editorYoutrack = ace.edit("editor-youtrack");
editorYoutrack.setTheme("ace/theme/monokai");
editorYoutrack.session.setMode("ace/mode/markdown");
editorYoutrack.setOptions({
    useSoftTabs: true,
    tabSize: 2,
    enableBasicAutocompletion: true,
    showPrintMargin: false,
    fixedWidthGutter: true,
    scrollPastEnd: 0.5,
    displayIndentGuides: true,
    fontSize: "medium",
    readOnly: true,
});
editorYoutrackTab.style.height = '100%';
editorYoutrackTab.style.width = '100%';

// Init editor with text
fetch("/file")
    .then((response) => response.json())
    .then((json) => initEditor(json.text, json.path));

function initEditor(text, path) {
    editor.setValue(text);
    editor.clearSelection()
    if (path.trim().length === 0) {
        saveTab.style.display = 'none';
        pathTab.style.display = 'none';
    } else {
        pathTab.innerHTML = path;
    }
}

// Hide logs
showLogs(false)

// Init tabs
activateTab('openapi')

// Set sevents 
buttons['swagger'].onclick = function() {activateTab('swagger')}
buttons['openapi'].onclick = function() {activateTab('openapi')}
buttons['youtrack'].onclick = function() {activateTab('youtrack')}

generateTab.onclick = function() {
    if (generated) {
        return
    }
    fetch('/generate', {
        method: 'POST',
        body: JSON.stringify({text:editor.getValue()}),
    })
        .then((response) => {
            if (response.ok) {
                return response.json();
            }
            return Promise.reject(response);
        })
        .then((json) => handleGeneration(json))
        .catch((response) => {
            console.log(response.status, response.statusText);
            response.text().then((text) => {
                handleGeneration({error: text})
            })
        });
};

saveTab.onclick = function() {
    var doSave = confirm('Перезаписать файл ' + pathTab.innerHTML + ' ?')
    if (!doSave) {
        return
    }

    fetch('/save', {
        method: 'POST',
        body: JSON.stringify({text:editor.getValue()}),
    })
        .then((response) => {
            if (response.ok) {
                console.log('saved')
                return
            }
            return Promise.reject(response);
        })
        .catch((response) => {
            console.log(response.status, response.statusText);
        });
};

function handleGeneration(data) {
    if (data.error.length !== 0) {
        showLogs(true)
        logsTab.innerHTML = 'Error: ' + data.error
    } else {
        showLogs(false)
        logsTab.innerHTML = 'No errors'
        showGenerate(false);
        editorOAPI.setValue(data.openapi);
        editorOAPI.clearSelection()
        editorYoutrack.setValue(data.youtrack);
        editorYoutrack.clearSelection()
        editorSwaggerTab.src = '/swagger/' + data.swagger_id + '/';
    }
}

function showGenerate(flag) {
    if (flag) {
        generateTab.style.background = 'linear-gradient(0deg, rgba(68,126,66,1) 13%, rgba(104,224,96,1) 100%)';
        generateTab.style.cursor = 'pointer';
    } else {
        generateTab.style.background = 'linear-gradient(0deg, rgba(135,138,135,1) 13%, rgb(221,221,221,1) 100%)';
        generateTab.style.cursor = 'default';
    }
    generated = !flag;
}

function showLogs(flag) {
    if (flag) {
        mainTab.style.gridTemplateRows = '0.2fr 2.2fr 0.3fr'
        logsTab.style.display = 'inline'
    } else {
        mainTab.style.gridTemplateRows = '0.177777fr 2.2fr 0fr'
        logsTab.style.display = 'none'
    }
};

function activateTab(tab) {
    if (tab === 'swagger') {
        editorSwaggerTab.style.height = '100%';
        editorSwaggerTab.style.width = '100%';
    } else {
        editorSwaggerTab.style.height = '0';
        editorSwaggerTab.style.width = '0';
    }

    if (tab === 'openapi') {
        editorOAPITab.style.height = '100%';
        editorOAPITab.style.width = '100%';
    } else {
        editorOAPITab.style.height = '0';
        editorOAPITab.style.width = '0';
    }

    if (tab === 'youtrack') {
        editorYoutrackTab.style.height = '100%';
        editorYoutrackTab.style.width = '100%';
    } else {
        editorYoutrackTab.style.height = '0';
        editorYoutrackTab.style.width = '0';
    }

    for (const [key, value] of Object.entries(buttons)) {
        if (key === tab) {
            value.style.backgroundColor = '#ccc';
            value.style.color = '#444';
        } else {
            value.style.backgroundColor = '#666';
            value.style.color = '#fff';
        }
    }
}



// swaggerBtn.onmouseover = function() {
//     this.style.backgroundColor = 'red';
// };
// swaggerBtn.onmouseout = function() {
//     this.style.backgroundColor = '';
// };
