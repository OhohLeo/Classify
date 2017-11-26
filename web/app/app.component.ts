import { Component, NgZone, OnInit, ViewChild } from '@angular/core'

import { ApiService, CollectionStatus, Event } from './api.service'
import { ImportsService } from './imports/imports.service'
import { BufferService } from './buffer/buffer.service'
import { CollectionsService } from './collections/collections.service'

import { Collection } from './collections/collection'

import { ListCollectionsComponent } from './collections/list.component'
import { BufferComponent } from './buffer/buffer.component'
import { BufferItemComponent } from './buffer/item.component'
import { BufferItem } from './buffer/item'
import { Item } from './collections/item'

declare var jQuery: any;

enum AppStatus {
    NONE = 0,
    COLLECTION,
    IMPORTS,
    EXPORTS,
    CONFIGS,
    BUFFER_ITEM,
}

@Component({
    selector: 'app',
    templateUrl: './app.component.html',
})

export class AppComponent implements OnInit {
    @ViewChild(ListCollectionsComponent) collections: ListCollectionsComponent
    @ViewChild(BufferComponent) buffer: BufferComponent

    public appStatus = AppStatus
    public status = AppStatus.NONE

    public title = "Classify"

    public collection: Collection

    public modalTitle: string
    public modalMsg: string

    private importsLoop: any
    private importsRunningNb: number

    @ViewChild(BufferItemComponent) bufferItemComponent: BufferItemComponent
    public bufferItem: BufferItem

    constructor(private zone: NgZone,
        private apiService: ApiService,
        private importsService: ImportsService,
        private bufferService: BufferService,
        private collectionsService: CollectionsService) { }

    ngOnInit() {

        // Initialisation de la side bar
        jQuery(".button-collapse").sideNav();

        // Logo d'importation
        this.importsLoop = jQuery("i#imports-loop")
        this.importsRunningNb = 0;

        // Inscription au flux
        this.apiService.getStream()
            .subscribe((e: Event) => {

                console.log("EVENT!", e)

                // Detect restart application
                if (e.event === "start") {
                    window.location.replace("/");
                    return;
                }

                // Import data
                if (new RegExp("^import").test(e.event)) {
                    this.handleImport(e);
                    return;
                }

                // Collections reception
                if (new RegExp("^collection").test(e.event)) {
                    this.handleCollection(e);
                    return;
                }
            })

        this.apiService.subscribeCollectionChange(
            (collection: Collection, status: CollectionStatus) => {

                if (collection === undefined) {
                    this.onChangeCollection()
                    return
                }

                this.title = collection.name
                this.collection = collection

                switch (status) {
                    case CollectionStatus.CREATED:
                    case CollectionStatus.MODIFIED:
                    case CollectionStatus.SELECTED:
                        this.onCollection()
                        break;
                    case CollectionStatus.DELETED:
                        this.status = AppStatus.NONE
                        break
                }
            })
    }

    onCollection() {
        this.onNewState(AppStatus.COLLECTION)
    }

    onImports() {
        this.onNewState(AppStatus.IMPORTS)
    }

    onExports() {
        this.onNewState(AppStatus.EXPORTS)
    }

    onConfigs() {
        this.onNewState(AppStatus.CONFIGS)
    }

    onBufferItem(bufferItem: BufferItem) {
        this.zone.run(() => {
            this.bufferItem = bufferItem
            this.onNewState(AppStatus.BUFFER_ITEM)
        })
    }

    onBuffer() {
        this.buffer.start()
    }

    onNewState(nextStatus: AppStatus) {
        this.resetCollectionState()
        this.status = nextStatus
    }

    // Gestion des nouveaux imports
    handleImport(e: Event) {

        console.log("IMPORT?", e)

        // Send notifications to the imports list
        this.importsService.addEvent(e);

        // Display imports status
        if (new RegExp('status$').test(e.event)) {

            console.log("Status??")

            // Status 'TRUE': rotate refresh logo
            if (e.data) {
                this.importsLoop.addClass("rotation")
                this.importsRunningNb++
            }
            // Status 'FALSE'
            else if (this.importsRunningNb > 0) {
                this.importsRunningNb--
            }

            // No more imports : stop logo rotation
            if (this.importsRunningNb < 1) {
                this.importsLoop.removeClass("rotation")
            }
        }
    }

    handleCollection(e: Event) {

        let names = e.event.split("/");
        let size = names.length;

        if (size <= 1) {
            console.error("Invalid collection event '" + e.event + "'")
            return;
        }

        let collection = names[1];
        let destination: string
        if (size > 2) {
            destination = names[2];
        }

        // Send notifications to the imports list
        switch (destination) {
            case "buffer":

                let bufferItem = new BufferItem(e.data)

                if (this.bufferItem != undefined && this.bufferItem.id == bufferItem.id) {
                    this.bufferItemComponent.onUpdate(bufferItem)
                    this.bufferItem = bufferItem
                }

                this.bufferService.addEvent(collection, e, bufferItem)
                break;
            case "items":
                this.collectionsService.addEvent(collection, e, new Item(e))
                break;
        }
    }

    onError(title: string, msg: string) {

        console.error(title + "error :" + msg)

        this.modalTitle = title + " error!"
        this.modalMsg = msg
        jQuery('#modal').openModal()
    }

    stopModal() {
        jQuery('#modal').closeModal()
    }

    // Affiche la création d'une nouvelle collection
    onNewCollection() {
        if (this.collections) {
            this.collections.onNewCollection()
        }
    }

    // Affiche la liste des collections à sélectionner
    onChangeCollection() {
        if (this.collections
            && this.collections.onChooseCollection(undefined)) {
            this.status = AppStatus.NONE
        }
    }

    // Affiche la liste des collections à sélectionner si aucune
    // collection n'est actuellement sélectionnée
    resetCollectionState() {
        if (this.collections)
            this.collections.resetCollectionState()
    }
}
