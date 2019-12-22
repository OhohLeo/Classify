import { Injectable } from '@angular/core';
import { Response } from '@angular/http';

import { Observable } from 'rxjs/Rx';

import { ApiService, Event } from '../api.service';
import { BufferService } from '../buffer/buffer.service';
import { ReferencesService } from '../references/references.service'

import { Convert2Imap } from './imap/imap'
import { Convert2Directory } from './directory/directory'

import { BaseElement } from '../base'

@Injectable()
export class ImportsService {

    enableCache: boolean
    imports: Map<string, BaseElement[]> = new Map<string, BaseElement[]>()
    importsByName: Map<string, BaseElement> = new Map<string, BaseElement>()
    updateList: any

    private eventObservers = {}
    private convertToImport: { [index:string]: (string, any) => BaseElement } = {
	"imap": Convert2Imap,
	"directory": Convert2Directory,
    }

    constructor(private apiService: ApiService,
		private bufferService: BufferService,
		private referencesService: ReferencesService) {}
    
    // Set update import list function
    setUpdateList(updateList: any) {
        this.updateList = updateList;
    }

    // Refresh the import list
    private update() {
        if (this.updateList != undefined)
            this.updateList()
    }

    // Check if import does exist
    hasImport(search: BaseElement): boolean {
        return this.hasSameImportName(search.name)
    }

    // Check if import does exist
    hasSameImportName(name: string): boolean {
        return this.importsByName.get(name) != undefined
    }

    // Check if import does exist
    hasSameImport(search: BaseElement): boolean {
        let imports = this.imports.get(search.getRef())
        if (imports === undefined) {
            return false
        }

        for (let i of imports) {
            if (i.compare(search)) {
                return true
            }
        }

        return false
    }

    private add(i: BaseElement) {

        // Store imports by name
        this.importsByName.set(i.name, i)

        // Store imports by ref
        if (this.imports.get(i.getRef()) === undefined) {
            this.imports.set(i.getRef(), [])
        }

        this.imports.get(i.getRef()).push(i)
    }

    addImport(i: BaseElement, onParams: any, onSuccess: any) {

        // Disable cache
        this.enableCache = false

        if (this.hasSameImport(i)) {
            console.error("Already existing " + i.getRef())
            return
        }

	let name = i.getID()
	if (this.hasSameImportName(name)) {
	    console.error("Already existing name " + name)
            return
	}

        return this.apiService.post(
            "imports", {
                "name": name,
                "ref": i.getRef(),
                "params": i.getParams(),
                "collections": [this.apiService.getCollectionName()],
            })
            .subscribe(rsp => {

                if (rsp.status != 200) {
                    throw new Error('Error when adding new import: ' + rsp.status);
                }

                let body = rsp.json()

                if (body === undefined || body.name === undefined) {

                    if (onParams !== undefined && onParams(body))
                        return

                    throw new Error('Name not found when adding new import!')
                }

                this.add(i)

                this.update()

                if (onSuccess !== undefined) {
                    onSuccess(i)
                }
            })
    }

    private delete(i: BaseElement) {

        // Delete import by name
        this.importsByName.delete(i.name)

        // Delete import by ref
        let importList = this.imports.get(i.getRef())
        for (let idx in importList) {
            let importItem = importList[idx]
            if (importItem.name === i.name) {
                importList.splice(+idx, 1)
                break;
            }
        }

        // Remove import refs with no imports
        if (importList.length == 0) {
            this.imports.delete(i.getRef())
        }
    }

    deleteImport(i: BaseElement) {

        // Disable cache
        this.enableCache = false

        if (this.hasImport(i) === false) {
            console.error("No existing " + i.getRef())
            return
        }

        let urlParams = "?name=" + i.name
            + "&collection=" + this.apiService.getCollectionName();

        return this.apiService.delete("imports" + urlParams)
            .subscribe(rsp => {

                if (rsp.status != 204) {
                    throw new Error('Error when deleting import: ' + rsp.status)
                }

                // Delete import
                this.delete(i)

                this.update()
            })
    }

    getUrl(i: BaseElement): string {
	return "imports/" + i.name
    }

    startImport(i: BaseElement) {
        return this.actionImport(true, i)
    }

    stopImport(i: BaseElement) {
        return this.actionImport(false, i)
    }

    actionImport(start: boolean, i: BaseElement) {

        if (this.hasImport(i) === false) {
            console.error("No existing " + i.getRef())
            return
        }

        let action = start ? "start" : "stop"
        let urlParams = "?name=" + i.name
            + "&collection=" + this.apiService.getCollectionName();

        return this.apiService.put("imports/" + action + urlParams)
            .subscribe(rsp => {
                if (rsp.status != 204) {
                    throw new Error('Error when ' + action + ' import: ' + rsp.status)
                }

                if (start)
                    this.bufferService.disableCache();
            })
    }

    // Ask for current imports list
    getImports() {
        return new Observable(observer => {

            // Returns the cache if the list should not have changed
            if (this.imports && this.enableCache === true) {
                observer.next(this.imports)
                return
            }

            // Ask for the current list
            this.apiService.get("imports").subscribe(rsp => {

                // Init the import lists
                this.imports = new Map<string, BaseElement[]>()
                this.importsByName = new Map<string, BaseElement>()

                for (let importName in rsp) {
		    let importRsp = rsp[importName]

		    let importRef = importRsp["ref"]
                    let convert = this.convertToImport[importRef]
                    if (convert === undefined) {
                        console.error(
                            "Unknown import ref '" + importRef + "'")
                        continue
                    }

                    let i = convert(importName, importRsp["params"])
                    if (i === undefined)
                        continue

                    this.add(i)
                }

                this.enableCache = true

                observer.next(this.imports)
            })
        })
    }

    subscribeEvents(name: string): Observable<Event> {

        if (this.eventObservers[name] != undefined) {
            console.error("Already existing observer", name)
            return;
        }

        return Observable.create(observer => {

            // Initialisation de l'observer
            this.eventObservers[name] = observer

            return () => delete this.eventObservers[name]
        })
    }

    addEvent(event: Event) {
        for (let name in this.eventObservers) {
            this.eventObservers[name].next(event)
        }
    }
}
