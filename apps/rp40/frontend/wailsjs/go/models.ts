export namespace main {
	
	export class AssessmentView {
	    Recommendation: string;
	    KSource: string;
	    BSource: string;
	    RecommendedKMD: number;
	    RecommendedBPsi: number;
	    Rationale: string[];
	    B22Valid: boolean;
	    B22Marginal: boolean;
	    FullBCollapsed: boolean;
	    HighPerm: boolean;
	    RelativeKGap: number;
	    MeanPmPsi: number;
	
	    static createFrom(source: any = {}) {
	        return new AssessmentView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Recommendation = source["Recommendation"];
	        this.KSource = source["KSource"];
	        this.BSource = source["BSource"];
	        this.RecommendedKMD = source["RecommendedKMD"];
	        this.RecommendedBPsi = source["RecommendedBPsi"];
	        this.Rationale = source["Rationale"];
	        this.B22Valid = source["B22Valid"];
	        this.B22Marginal = source["B22Marginal"];
	        this.FullBCollapsed = source["FullBCollapsed"];
	        this.HighPerm = source["HighPerm"];
	        this.RelativeKGap = source["RelativeKGap"];
	        this.MeanPmPsi = source["MeanPmPsi"];
	    }
	}
	export class B22LinearityPoint {
	    MeanPressurePa: number;
	    MeanPressurePsi: number;
	    Response: number;
	    FitResponse: number;
	    Residual: number;
	
	    static createFrom(source: any = {}) {
	        return new B22LinearityPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MeanPressurePa = source["MeanPressurePa"];
	        this.MeanPressurePsi = source["MeanPressurePsi"];
	        this.Response = source["Response"];
	        this.FitResponse = source["FitResponse"];
	        this.Residual = source["Residual"];
	    }
	}
	export class B22View {
	    KMD: number;
	    BPsi: number;
	    R2: number;
	    Slope: number;
	    Intercept: number;
	    MaxResidual: number;
	    Linearity: B22LinearityPoint[];
	
	    static createFrom(source: any = {}) {
	        return new B22View(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.KMD = source["KMD"];
	        this.BPsi = source["BPsi"];
	        this.R2 = source["R2"];
	        this.Slope = source["Slope"];
	        this.Intercept = source["Intercept"];
	        this.MaxResidual = source["MaxResidual"];
	        this.Linearity = this.convertValues(source["Linearity"], B22LinearityPoint);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FullFitView {
	    KMD: number;
	    BPa: number;
	    BPsi: number;
	    SE: number;
	    R2: number;
	    A1: number;
	    A2: number;
	
	    static createFrom(source: any = {}) {
	        return new FullFitView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.KMD = source["KMD"];
	        this.BPa = source["BPa"];
	        this.BPsi = source["BPsi"];
	        this.SE = source["SE"];
	        this.R2 = source["R2"];
	        this.A1 = source["A1"];
	        this.A2 = source["A2"];
	    }
	}
	export class ChartPoint {
	    X: number;
	    Y: number;
	
	    static createFrom(source: any = {}) {
	        return new ChartPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	    }
	}
	export class CalculationView {
	    HasResult: boolean;
	    ProcessedRows: number;
	    DisplayKMD: number;
	    DisplayBPsi: number;
	    DisplayMethod: string;
	    DisplayWarnings: string[];
	    ProcessedCurve: ChartPoint[];
	    B22FitCurve: ChartPoint[];
	    Full: FullFitView;
	    B22: B22View;
	    Assessment: AssessmentView;
	
	    static createFrom(source: any = {}) {
	        return new CalculationView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.HasResult = source["HasResult"];
	        this.ProcessedRows = source["ProcessedRows"];
	        this.DisplayKMD = source["DisplayKMD"];
	        this.DisplayBPsi = source["DisplayBPsi"];
	        this.DisplayMethod = source["DisplayMethod"];
	        this.DisplayWarnings = source["DisplayWarnings"];
	        this.ProcessedCurve = this.convertValues(source["ProcessedCurve"], ChartPoint);
	        this.B22FitCurve = this.convertValues(source["B22FitCurve"], ChartPoint);
	        this.Full = this.convertValues(source["Full"], FullFitView);
	        this.B22 = this.convertValues(source["B22"], B22View);
	        this.Assessment = this.convertValues(source["Assessment"], AssessmentView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class LogEntry {
	    Time: string;
	    Level: string;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.Level = source["Level"];
	        this.Message = source["Message"];
	    }
	}
	export class PassportSampleOption {
	    ID: string;
	    Label: string;
	    Occurrences: number;
	    UpdatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new PassportSampleOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Label = source["Label"];
	        this.Occurrences = source["Occurrences"];
	        this.UpdatedAt = source["UpdatedAt"];
	    }
	}
	export class PassportSampleRecord {
	    SampleID: string;
	    Timestamp: string;
	    Gas: string;
	    LengthMM: number;
	    DiameterMM: number;
	    HeightMM: number;
	    VolumeCM3: number;
	    PorosityPct: number;
	    TemperatureC: number;
	    AtmosphericKPa: number;
	    CompressionMPa: number;
	    PoreVolumeCM3: number;
	    KGasMD: number;
	    KKlinkenbergMD: number;
	    BPsi: number;
	    RawRow: number;
	    SourceOccurrences: number;
	
	    static createFrom(source: any = {}) {
	        return new PassportSampleRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SampleID = source["SampleID"];
	        this.Timestamp = source["Timestamp"];
	        this.Gas = source["Gas"];
	        this.LengthMM = source["LengthMM"];
	        this.DiameterMM = source["DiameterMM"];
	        this.HeightMM = source["HeightMM"];
	        this.VolumeCM3 = source["VolumeCM3"];
	        this.PorosityPct = source["PorosityPct"];
	        this.TemperatureC = source["TemperatureC"];
	        this.AtmosphericKPa = source["AtmosphericKPa"];
	        this.CompressionMPa = source["CompressionMPa"];
	        this.PoreVolumeCM3 = source["PoreVolumeCM3"];
	        this.KGasMD = source["KGasMD"];
	        this.KKlinkenbergMD = source["KKlinkenbergMD"];
	        this.BPsi = source["BPsi"];
	        this.RawRow = source["RawRow"];
	        this.SourceOccurrences = source["SourceOccurrences"];
	    }
	}
	export class RP40Inputs {
	    SampleID: string;
	    Gas: string;
	    DiameterMM: number;
	    LengthMM: number;
	    PorosityPct: number;
	    TemperatureC: number;
	    AtmosphericKPa: number;
	    ReservoirMode: string;
	    ReservoirVolumeML: number;
	    CustomReservoirML: number;
	    LowPerm: boolean;
	    MinDeltaP: number;
	    MinDeltaT: number;
	    StartAt: number;
	
	    static createFrom(source: any = {}) {
	        return new RP40Inputs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SampleID = source["SampleID"];
	        this.Gas = source["Gas"];
	        this.DiameterMM = source["DiameterMM"];
	        this.LengthMM = source["LengthMM"];
	        this.PorosityPct = source["PorosityPct"];
	        this.TemperatureC = source["TemperatureC"];
	        this.AtmosphericKPa = source["AtmosphericKPa"];
	        this.ReservoirMode = source["ReservoirMode"];
	        this.ReservoirVolumeML = source["ReservoirVolumeML"];
	        this.CustomReservoirML = source["CustomReservoirML"];
	        this.LowPerm = source["LowPerm"];
	        this.MinDeltaP = source["MinDeltaP"];
	        this.MinDeltaT = source["MinDeltaT"];
	        this.StartAt = source["StartAt"];
	    }
	}
	export class RP40Snapshot {
	    PassportFilePath: string;
	    MeasurementFilePath: string;
	    Samples: PassportSampleOption[];
	    SelectedSample: PassportSampleRecord;
	    Inputs: RP40Inputs;
	    Calculation: CalculationView;
	    Logs: LogEntry[];
	    LastError: string;
	
	    static createFrom(source: any = {}) {
	        return new RP40Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PassportFilePath = source["PassportFilePath"];
	        this.MeasurementFilePath = source["MeasurementFilePath"];
	        this.Samples = this.convertValues(source["Samples"], PassportSampleOption);
	        this.SelectedSample = this.convertValues(source["SelectedSample"], PassportSampleRecord);
	        this.Inputs = this.convertValues(source["Inputs"], RP40Inputs);
	        this.Calculation = this.convertValues(source["Calculation"], CalculationView);
	        this.Logs = this.convertValues(source["Logs"], LogEntry);
	        this.LastError = source["LastError"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

