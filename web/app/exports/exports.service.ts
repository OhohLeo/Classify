import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Rx';
import { ApiService, Event } from './../api.service';
import { BufferService } from './../buffer/buffer.service';
import { Response } from '@angular/http';

export class ExportBase {

    public isRunning: boolean

    constructor(private type: string, public id: string) { }

    setId(id: string) {
        this.id = id;
    }

    getId(): string {
        return this.id
    }

    getType(): string {
        if (this.type === undefined)
            throw new Error("attribute 'type' should be defined!")

        return this.type
    }

    getParams(): any {
        throw new Error("method 'getParams' should be defined!")
    }

    display(): string {
        throw new Error("method 'display' should be defined!")
    }

    compare(i: ExportBase): boolean {
        if (this.type === undefined)
            throw new Error("attribute 'type' should be defined!")

        if (this.type != i.getType())
            return false

        return true
    }
}

export class Directory extends ExportBase {

    constructor(public id: string,
        public path: string,
        public isRecursive: boolean) {

        super("directory", id);

        if (isRecursive === undefined) {
            this.isRecursive = false
        }
    }

    getParams(): any {
        return {
            "path": this.path,
            "is_recursive": this.isRecursive ? true : false
        }
    }

    display(): string {
        return this.path.concat(this.isRecursive == true ? "/**" : "")
    }

    compare(i: Directory): boolean {
        return super.compare(i)
            && this.path === i.path
            && this.isRecursive == i.isRecursive
    }
}
@Injectable()
export class ExportsService {

    enableCache: boolean
    exports: Map<string, ExportBase[]> = new Map<string, ExportBase[]>()
    exportsById: Map<string, ExportBase> = new Map<string, ExportBase>()
    configs: any
    updateList: any

    private eventObservers = {}

    private convertToExport = {};


    constructor(private apiService: ApiService,
        private bufferService: BufferService) { }

    // Set update export list function
    setUpdateList(updateList: any) {
        this.updateList = updateList;
    }

    // Refresh the export list
    private update() {
        if (this.updateList != undefined)
            this.updateList()
    }

    // Check if export does exist
    hasExport(search: ExportBase): boolean {
        return this.exportsById.get(search.id) != undefined
    }

    // Check if export does exist
    hasSameExport(search: ExportBase): boolean {
        let exports = this.exports.get(search.getType())
        if (exports === undefined) {
            return false
        }

        for (let i of exports) {
            if (i.compare(search)) {
                return true
            }
        }

        return false
    }

    private add(i: ExportBase) {

        // Store exports by id
        this.exportsById.set(i.id, i)

        // Store exports by type
        if (this.exports.get(i.getType()) === undefined) {
            this.exports.set(i.getType(), [])
        }

        this.exports.get(i.getType()).push(i)
    }

    addExport(i: ExportBase) {

        // Disable cache
        this.enableCache = false

        if (this.hasSameExport(i)) {
            console.error("Already existing " + i.getType())
            return
        }

        return this.apiService.post(
            "exports", {
                "type": i.getType(),
                "params": i.getParams(),
                "collections": [this.apiService.getCollectionName()],
            })
            .subscribe(rsp => {

                if (rsp.status != 200) {
                    throw new Error('Error when adding new export: ' + rsp.status);
                }

                let body = rsp.json()

                if (body === undefined && body.id === undefined) {
                    throw new Error('Id not found when adding new export!');
                }

                i.setId(body.id)

                this.add(i)

                this.update()
            })
    }

    private delete(i: ExportBase) {

        // Delete export by id
        this.exportsById.delete(i.id)

        // Delete export by type
        let exportList = this.exports.get(i.getType())
        for (let idx in exportList) {
            let exportItem = exportList[idx]
            if (exportItem.id === i.getId()) {
                exportList.splice(+idx, 1)
                break;
            }
        }

        // Remove export types with no exports
        if (exportList.length == 0) {
            this.exports.delete(i.getType())
        }
    }

    deleteExport(i: ExportBase) {

        // Disable cache
        this.enableCache = false

        if (this.hasExport(i) === false) {
            console.error("No existing " + i.getType())
            return
        }

        let urlParams = "?id=" + i.getId()
            + "&collection=" + this.apiService.getCollectionName();

        return this.apiService.delete("exports" + urlParams)
            .subscribe(rsp => {

                if (rsp.status != 204) {
                    throw new Error('Error when deleting export: ' + rsp.status)
                }

                // Delete export
                this.delete(i)

                this.update()
            })
    }

    startExport(i: ExportBase) {
        return this.actionExport(true, i)
    }

    stopExport(i: ExportBase) {
        return this.actionExport(false, i)
    }

    actionExport(isStart: boolean, i: ExportBase) {

        if (this.hasExport(i) === false) {
            console.error("No existing " + i.getType())
            return
        }

        let action = isStart ? "start" : "stop"
        let urlParams = "?id=" + i.getId()
            + "&collection=" + this.apiService.getCollectionName();

        return this.apiService.put("exports/" + action + urlParams)
            .subscribe(rsp => {
                if (rsp.status != 204) {
                    throw new Error('Error when '
                        + action + ' export: ' + rsp.status)
                }

                if (isStart)
                    this.bufferService.disableCache();
            })
    }

    // Ask for current exports list
    getExports() {
        return new Observable(observer => {

            // Returns the cache if the list should not have changed
            if (this.exports && this.enableCache === true) {
                observer.next(this.exports)
                return
            }

            // Ask for the current list
            this.apiService.get("exports").subscribe(rsp => {

                // Init the export lists
                this.exports = new Map<string, ExportBase[]>()
                this.exportsById = new Map<string, ExportBase>()

                for (let exportType in rsp) {

                    let convert = this.convertToExport[exportType]
                    if (convert === undefined) {
                        console.error(
                            "Unknown export type '" + exportType + "'")
                        continue
                    }

                    for (let exportId in rsp[exportType]) {
                        let i = convert(exportId, rsp[exportType][exportId])
                        if (i === undefined)
                            continue

                        this.add(i)
                    }
                }

                this.enableCache = true

                observer.next(this.exports)
            })
        })
    }

    // Ask for current export config list
    getExportsConfig(exportType: string) {
        return new Observable(observer => {

            // Export config list should not change a lot
            if (this.configs) {
                observer.next(this.configs[exportType])
                return
            }

            // Ask for the current export config list
            return this.apiService.get("exports/config")
                .subscribe(rsp => {

                    // Store as cache the current export config list
                    this.configs = rsp

                    // Return the export config list
                    observer.next(rsp[exportType])
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